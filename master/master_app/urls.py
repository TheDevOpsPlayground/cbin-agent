from django.urls import path
from django.contrib.auth import views as auth_views
from . import views
from .views import view_logs
from .views import data_viewer, open_directory, download_file
from django.urls import path
from django.views.generic import RedirectView

urlpatterns = [
    path('login/', auth_views.LoginView.as_view(template_name='login.html'), name='login'),
    path('dashboard/', views.dashboard, name='dashboard'),
    path('logout/', auth_views.LogoutView.as_view(), name='logout'),
    path('data-viewer/', views.data_viewer, name='data_viewer'),
    path('data-viewer/open/<path:subdir>/', views.open_directory, name='open_directory'),
    path('data-viewer/download/<path:filename>/', views.download_file, name='download_file'),
    path('view-logs/', view_logs, name='view_logs'),
    path('mount/', views.mount, name='mount'),
    path('', RedirectView.as_view(url='/login/', permanent=False)), 
    path('add/', views.add_agent, name='add_agent'),
    path('view/', views.view_agents, name='view_agents'),
    path('delete/<int:agent_id>/', views.delete_agent, name='delete_agent'),
    path('variables/', views.variables, name='variables'),

]

