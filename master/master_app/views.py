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

import subprocess
import shlex

def run_command(command):
   """
   Run a shell command with sudo, capturing output or error.
   """
   try:
       # Use subprocess to run command with sudo
       result = subprocess.run(
           ['sudo'] + shlex.split(command), 
           capture_output=True, 
           text=True, 
           check=True
       )
       return result.stdout.strip()
   except subprocess.CalledProcessError as e:
       return f"Error: {e.stderr.strip()}"
   except Exception as e:
       return f"Unexpected error: {str(e)}"
#############issue-1 #################
def check_export_exists(client_ip):
    """
    Check if the export rule for the given client IP exists in /etc/exports.
    """
    try:
        # Use run_command to read file contents with sudo
        exports_content = run_command('cat /etc/exports')
        
        # More robust export rule detection
        export_pattern = f"{BASE_DIR}/{client_ip}"
        
        # Check if any line contains the export pattern
        return any(
            export_pattern in line.strip() 
            for line in exports_content.splitlines() 
            if line.strip() and not line.strip().startswith('#')
        )
    except Exception:
        return False


#############################
def create_base_directory():
    """
    Create the base directory for NFS exports if it doesn't exist.
    """
    try:
        # Ensure BASE_DIR exists with proper permissions
        result = run_command(f"mkdir -p {BASE_DIR}")
        run_command(f"sudo chmod 755 {BASE_DIR}")
        return True
    except Exception as e:
        print(f"Base directory creation error: {str(e)}")
        return False
import os
import subprocess
import shlex
import logging

# Configure logging
logging.basicConfig(level=logging.DEBUG, 
                    format='%(asctime)s - %(levelname)s - %(message)s',
                    filename='/tmp/nfs_setup.log')

def run_command(command):
    """
    Run a shell command with comprehensive logging.
    """
    try:
        logging.info(f"Running command: {command}")
        result = subprocess.run(
            command, 
            shell=True, 
            capture_output=True, 
            text=True, 
            check=True
        )
        logging.info(f"Command output: {result.stdout}")
        return result.stdout.strip()
    except subprocess.CalledProcessError as e:
        logging.error(f"Command failed: {command}")
        logging.error(f"Error output: {e.stderr}")
        return f"Error: {e.stderr}"
    except Exception as e:
        logging.error(f"Unexpected error: {str(e)}")
        return f"Unexpected error: {str(e)}"

def setup_client_directory(client_ip, fsid):
    """
    Create a directory for the client IP and configure the NFS export.
    """
    try:
        # Use sudo to ensure proper permissions
        client_dir = os.path.join(BASE_DIR, client_ip)
        
        # Create directory with sudo
        run_command(f"sudo mkdir -p {client_dir}")
        run_command(f"sudo chmod 777 {client_dir}")

        # Prepare export rule
        export_rule = f"{client_dir} {client_ip}(rw,sync,no_subtree_check,all_squash,anonuid=65534,anongid=65534,fsid={fsid})"

        # Append to exports file using sudo
        append_cmd = f'sudo bash -c \'echo "{export_rule}" >> /etc/exports\''
        run_command(append_cmd)

        return True
    except Exception as e:
        print(f"Client directory setup error: {str(e)}")
        return False
def apply_nfs_exports():
    """
    Apply NFS exports and restart the NFS server with comprehensive error handling.
    """
    try:
        # Validate exports file
        run_command("exportfs -v")
        
        # Reload exports
        run_command("exportfs -ra")
        
        # Restart NFS service
        run_command("sudo systemctl restart nfs-kernel-server")
        
        return True
    except Exception as e:
        print(f"NFS exports application error: {str(e)}")
        return False


import re
import ipaddress

def is_valid_ip(ip_address):
    """
    Validate if the given string is a valid IP address.
    
    Args:
        ip_address (str): IP address to validate
    
    Returns:
        bool: True if valid IP, False otherwise
    """
    try:
        # Try parsing as IPv4
        ipaddress.IPv4Address(ip_address)
        
        # Additional regex check for format
        ip_pattern = r'^(\d{1,3}\.){3}\d{1,3}$'
        if not re.match(ip_pattern, ip_address):
            return False
        
        # Validate each octet is between 0-255
        octets = ip_address.split('.')
        return all(0 <= int(octet) <= 255 for octet in octets)
    
    except (ipaddress.AddressValueError, ValueError):
        return False











def add_agent(request):
    """
    Handle the addition of a new agent with detailed error handling.
    """
    if request.method == "POST":
        ip_address = request.POST.get('ip_address', '').strip()
        
        # First, validate IP address
        if not is_valid_ip(ip_address):
            messages.error(request, f"Invalid IP address: {ip_address}")
            return render(request, 'add_agent.html', {'form': AddAgentForm()})
        if ip_address:
            try:
                # Check if agent exists
                if Agent.objects.filter(ip_address=ip_address).exists():
                    messages.error(request, f"Agent with IP {ip_address} already exists.")
                    return render(request, 'add_agent.html', {'form': AddAgentForm()})

                # Check if export exists
                if check_export_exists(ip_address):
                    messages.error(request, f"Export rule for IP {ip_address} already exists.")
                    return render(request, 'add_agent.html', {'form': AddAgentForm()})

                # Perform setup steps
                if not create_base_directory():
                    messages.error(request, 'Failed to create base directory.')
                    return render(request, 'add_agent.html', {'form': AddAgentForm()})

                if not setup_client_directory(ip_address, fsid=1):
                    messages.error(request, 'Failed to setup client directory.')
                    return render(request, 'add_agent.html', {'form': AddAgentForm()})

                if not apply_nfs_exports():
                    messages.error(request, 'Failed to apply NFS exports.')
                    return render(request, 'add_agent.html', {'form': AddAgentForm()})

                # If all steps succeed, save the agent
                agent = Agent(ip_address=ip_address, status=0)
                agent.save()
                messages.success(request, f"Agent {ip_address} added successfully.")
                return redirect('view_agents')

            except Exception as e:
                messages.error(request, f'Unexpected error: {str(e)}')
                print(f"Unexpected error in add_agent: {str(e)}")

    return render(request, 'add_agent.html', {'form': AddAgentForm()})





























def view_agents(request):
    """
    Render the view_agents template with all the agents.
    """
    agents = Agent.objects.all()
    return render(request, 'view_agents.html', {'agents': agents})





#######################



def remove_export_rule(client_ip):
   """
   Remove the export rule for the given client IP from /etc/exports.
   """
   try:
       export_rule = f"{BASE_DIR}/{client_ip} {client_ip}("
       
       # Use sudo to read and modify /etc/exports
       read_cmd = f"sudo grep -v '{export_rule}' /etc/exports"
       exports_content = run_command(read_cmd)
       
       # Write back filtered content using sudo
       write_cmd = f"sudo tee /etc/exports > /dev/null << EOF\n{exports_content}\nEOF"
       run_command(write_cmd)
       
       # Restart NFS server
       run_command("sudo systemctl restart nfs-kernel-server")
       
       return True
   except Exception as e:
       print(f"Error removing export rule: {str(e)}")
       return False

def delete_agent(request, agent_id):
   """
   Handle the deletion of an agent and its corresponding NFS export rule.
   """
   try:
       agent = Agent.objects.get(id=agent_id)
       ip_address = agent.ip_address
       
       if remove_export_rule(ip_address):
           agent.delete()
           messages.success(request, 'Agent and corresponding export rule deleted successfully.')
       else:
           messages.error(request, 'Failed to remove export rule or restart NFS server.')
   except Agent.DoesNotExist:
       messages.error(request, 'Agent not found.')
   
   return redirect('view_agents')