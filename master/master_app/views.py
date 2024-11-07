import os
import subprocess
from django.contrib.auth.decorators import login_required
from django.shortcuts import render, redirect
from django.http import FileResponse, Http404
from django.contrib import messages
from .models import Agent
from .forms import AddAgentForm
from urllib.parse import unquote
import environ

# Initialize environment variables
env = environ.Env()
environ.Env.read_env()

# Base directory where files are stored
BASE_DIR = os.getenv('BASE_PATH', '/mnt/s3-bucket')

@login_required
def dashboard(request):
    """
    Render the dashboard template for authenticated users.
    """
    return render(request, 'dashboard.html')

def view_logs(request):
    """
    Display log files based on a search query.
    """
    all_logs = {}
    search_query = request.GET.get('search', '').strip()

    # Check if the base directory exists
    if os.path.exists(BASE_DIR):
        # Loop through directories in the base directory
        for foldername in os.listdir(BASE_DIR):
            folder_path = os.path.join(BASE_DIR, foldername)
            if os.path.isdir(folder_path):
                # Check if the log.txt file exists in the directory
                log_file_path = os.path.join(folder_path, 'log.txt')
                if os.path.exists(log_file_path):
                    with open(log_file_path, 'r') as log_file:
                        logs = log_file.read()
                        if search_query:
                            # Filter the logs based on the search query
                            logs = filter_logs(logs, search_query)
                        all_logs[foldername] = logs

    return render(request, 'view_logs.html', {'all_logs': all_logs, 'search_query': search_query})

def filter_logs(logs, search_query):
    """
    Filter log entries by the search query (case-insensitive).
    """
    filtered_logs = [line for line in logs.splitlines() if search_query.lower() in line.lower()]
    return '\n'.join(filtered_logs)

def data_viewer(request):
    """
    Render the data viewer page, starting from the root directory.
    """
    return open_directory(request, '')

def open_directory(request, subdir):
    """
    Open a directory and list its contents.
    """
    subdir = unquote(subdir)
    directory_path = safe_join(BASE_DIR, subdir)

    if not os.path.exists(directory_path):
        raise Http404("Directory does not exist")

    # Get the list of directories and files in the current directory
    dirs = [entry.name for entry in os.scandir(directory_path) if entry.is_dir()]
    files = [entry.name for entry in os.scandir(directory_path) if not entry.is_dir()]

    context = {
        'current_dir': subdir,
        'dirs': dirs,
        'files': files
    }

    return render(request, 'data_viewer.html', context)

def safe_join(base_path, *paths):
    """
    Safely join the base path and the provided paths, ensuring the final path is within the base path.
    """
    final_path = os.path.join(base_path, *paths)
    if not final_path.startswith(base_path):
        raise Http404("Access Denied")
    return final_path

def download_file(request, filename):
    """
    Download a file from the base directory.
    """
    filename = unquote(filename)
    file_path = safe_join(BASE_DIR, filename)

    if os.path.exists(file_path):
        return FileResponse(open(file_path, 'rb'), as_attachment=True, filename=filename)
    else:
        raise Http404("File does not exist")

def mount(request):
    """
    Render the mount template with the base path.
    """
    base_path = os.getenv('BASE_PATH', 'Not Available')
    return render(request, 'mount.html', {'BASE_PATH': base_path})

def variables(request):
    """
    Render the variables template with all environment variables.
    """
    env_vars = {key: os.getenv(key) for key in os.environ.keys()}
    return render(request, 'variables.html', {'env_vars': env_vars})

def run_command(command):
    """
    Run a shell command and handle errors, returning the output or error message.
    """
    try:
        result = subprocess.run(command, check=True, shell=True, capture_output=True, text=True)
        return result.stdout
    except subprocess.CalledProcessError as e:
        return f"Error running command: {command}\n{e.stderr}"

def check_export_exists(client_ip):
    """
    Check if the export rule for the given client IP exists in /etc/exports.
    """
    export_rule = f"{BASE_DIR}/{client_ip} {client_ip}("
    exports_file_path = "/etc/exports"
    try:
        with open(exports_file_path, "r") as exports_file:
            return any(line.strip().startswith(export_rule) for line in exports_file)
    except IOError:
        return False

def create_base_directory():
    """
    Create the base directory for NFS exports if it doesn't exist.
    """
    if not os.path.exists(BASE_DIR):
        result = run_command(f"sudo mkdir -p {BASE_DIR}")
        return "Error" not in result

def setup_client_directory(client_ip, fsid):
    """
    Create a directory for the client IP and configure the NFS export.
    """
    client_dir = os.path.join(BASE_DIR, client_ip)
    if not os.path.exists(client_dir):
        result = run_command(f"sudo mkdir -p {client_dir}")
        if "Error" in result:
            return False
        run_command(f"sudo chmod 777 {client_dir}")

    export_rule = f"{client_dir} {client_ip}(rw,sync,no_subtree_check,all_squash,anonuid=65534,anongid=65534,fsid={fsid})"

    try:
        with open("/etc/exports", "a") as exports_file:
            if all(line.strip() != export_rule for line in exports_file):
                exports_file.write(f"{export_rule}\n")
    except IOError:
        return False

    return True

def apply_nfs_exports():
    """
    Apply NFS exports and restart the NFS server with error handling.
    """
    if "Error" in run_command("sudo exportfs -ua"):
        return False
    if "Error" in run_command("sudo exportfs -ra"):
        return False
    if "Error" in run_command("sudo systemctl restart nfs-kernel-server"):
        return False
    return True

def add_agent(request):
    """
    Handle the addition of a new agent.
    """
    if request.method == "POST":
        ip_address = request.POST.get('ip_address')
        if ip_address:
            # Check if the agent with the same IP address already exists
            if Agent.objects.filter(ip_address=ip_address).exists():
                messages.error(request, f"Agent with IP {ip_address} already exists.")
            # Check if the export rule for this IP already exists
            elif check_export_exists(ip_address):
                messages.error(request, f"Export rule for IP {ip_address} already exists.")
            # Create the base directory and client directory, then apply NFS exports
            elif create_base_directory() and setup_client_directory(ip_address, fsid=1):
                if apply_nfs_exports():
                    # Add the agent to the database
                    agent = Agent(ip_address=ip_address, status=0)
                    agent.save()
                    messages.success(request, f"Agent {ip_address} added successfully.")
                    return redirect('view_agents')
                else:
                    messages.error(request, 'Failed to set up NFS or require sudo access.')
            else:
                messages.error(request, 'Failed to set up NFS or require sudo access.')

    return render(request, 'add_agent.html', {'form': AddAgentForm()})

def view_agents(request):
    """
    Render the view_agents template with all the agents.
    """
    agents = Agent.objects.all()
    return render(request, 'view_agents.html', {'agents': agents})

def remove_export_rule(client_ip):
    """
    Remove the export rule for the given client IP from /etc/exports.
    """
    export_rule = f"{BASE_DIR}/{client_ip} {client_ip}("
    exports_file_path = "/etc/exports"
    try:
        with open(exports_file_path, "r") as exports_file:
            exports_content = [line for line in exports_file if not line.strip().startswith(export_rule)]
        with open(exports_file_path, "w") as exports_file:
            exports_file.writelines(exports_content)
        return True
    except IOError:
        return False

def delete_agent(request, agent_id):
    """
    Handle the deletion of an agent and its corresponding NFS export rule.
    """
    try:
        agent = Agent.objects.get(id=agent_id)
        ip_address = agent.ip_address
        if remove_export_rule(ip_address) and apply_nfs_exports():
            agent.delete()
            messages.success(request, 'Agent and corresponding export rule deleted successfully.')
        else:
            messages.error(request, 'Failed to remove export rule or require sudo access.')
    except Agent.DoesNotExist:
        messages.error(request, 'Agent not found.')
    return redirect('view_agents')