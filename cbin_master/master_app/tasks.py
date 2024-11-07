# recycle_bin/agent_manager/tasks.py
from celery import shared_task
from .models import Agent  # or from yourapp.models import Agent
import requests
@shared_task
def check_agent_health():
    agents = Agent.objects.all()  
    for agent in agents:
        try:
            response = requests.get(f'http://{agent.ip_address}:10001/health', timeout=5)
            agent.status = response.status_code == 200  # Update status based on response
        except requests.RequestException:
            agent.status = False  # Set to inactive on exception
        agent.save()
