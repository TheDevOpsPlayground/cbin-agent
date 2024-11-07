from pathlib import Path
import os
from dotenv import load_dotenv

# Load environment variables from .env file
load_dotenv()

# Fetch environment variables
SECRET_KEY = os.getenv('SECRET_KEY')
DEBUG = os.getenv('DEBUG') == 'True'  # Convert from string to boolean
ALLOWED_HOSTS = os.getenv('ALLOWED_HOSTS', '').split(',')

# Use Path for BASE_DIR to make path manipulations easier
BASE_DIR = Path(__file__).resolve().parent.parent

# Static files and templates settings
TEMPLATES = [
    {
        'BACKEND': 'django.template.backends.django.DjangoTemplates',
        'DIRS': [BASE_DIR / 'templates'],  # Corrected to use Path objects
        'APP_DIRS': True,
        'OPTIONS': {
            'context_processors': [
                'django.template.context_processors.debug',
                'django.template.context_processors.request',
                'django.contrib.auth.context_processors.auth',
                'django.contrib.messages.context_processors.messages',
            ],
        },
    },
]

MIDDLEWARE = [
    'django.middleware.security.SecurityMiddleware',
    'django.contrib.sessions.middleware.SessionMiddleware',  # Ensure this is before AuthenticationMiddleware
    'django.middleware.common.CommonMiddleware',
    'django.middleware.csrf.CsrfViewMiddleware',
    'django.contrib.auth.middleware.AuthenticationMiddleware',  # Should come after SessionMiddleware
    'django.contrib.messages.middleware.MessageMiddleware',  # Should be included
    'django.middleware.clickjacking.XFrameOptionsMiddleware',
]


STATIC_URL = 'static/'
STATICFILES_DIRS = [BASE_DIR / 'static']  # Corrected to use Path objects

# Database settings (example using SQLite, can be adjusted for PostgreSQL, etc.)
DATABASE_URL = os.getenv('DATABASE_URL', 'sqlite:///db.sqlite3')

DATABASES = {
    'default': {
        'ENGINE': 'django.db.backends.sqlite3',
        'NAME': BASE_DIR / 'db.sqlite3',  # Corrected to use Path objects
    }
}

# Celery settings
CELERY_BROKER_URL = os.getenv('CELERY_BROKER_URL', 'redis://localhost:6379/0')
CELERY_ACCEPT_CONTENT = ['json']
CELERY_TASK_SERIALIZER = 'json'

from celery.schedules import crontab
CELERY_BEAT_SCHEDULE = {
    'check-agent-health-every-minute': {
        'task': 'master_app.tasks.check_agent_health',  # Updated to your new app
        'schedule': crontab(minute='*'),
    },
}

INSTALLED_APPS = [
    'django.contrib.admin',
    'django.contrib.auth',
    'django.contrib.contenttypes',
    'django.contrib.sessions',
    'django.contrib.messages',
    'django.contrib.staticfiles',


    'master_app',  # Updated to 'master_app'
]

# Password validation settings
AUTH_PASSWORD_VALIDATORS = [
    {
        'NAME': 'django.contrib.auth.password_validation.UserAttributeSimilarityValidator',
    },
    {
        'NAME': 'django.contrib.auth.password_validation.MinimumLengthValidator',
    },
    {
        'NAME': 'django.contrib.auth.password_validation.CommonPasswordValidator',
    },
    {
        'NAME': 'django.contrib.auth.password_validation.NumericPasswordValidator',
    },
]

# Timezone settings
LANGUAGE_CODE = 'en-us'
TIME_ZONE = 'UTC'
USE_I18N = True
USE_TZ = True


#STATIC_URL = 'static/'
#STATICFILES_DIRS = [os.path.join(BASE_DIR, 'static')]

DEFAULT_AUTO_FIELD = 'django.db.models.BigAutoField'
LOGIN_REDIRECT_URL = '/dashboard/'
LOGOUT_REDIRECT_URL = '/login/'

# Load additional settings from environment variables
ALLOWED_HOSTS = os.getenv('ALLOWED_HOSTS', '127.0.0.1,localhost').split(',')


ROOT_URLCONF = 'master_app.urls'


# your_project/settings.py
CELERY_BROKER_URL = 'redis://localhost:6379/0'  # You can use RabbitMQ or Redis
CELERY_ACCEPT_CONTENT = ['json']
CELERY_TASK_SERIALIZER = 'json'

from celery.schedules import crontab

# recycle_bin/settings.py
CELERY_BEAT_SCHEDULE = {
    'check-agent-health-every-minute': {
        'task': 'master_app.tasks.check_agent_health',
        'schedule': crontab(minute='*'),
    },
}



# settings.py

LOGGING = {
    'version': 1,
    'disable_existing_loggers': False,
    'handlers': {
        'file': {
            'level': 'DEBUG',
            'class': 'logging.FileHandler',
            'filename': 'django.log',
        },
    },
    'loggers': {
        'django': {
            'handlers': ['file'],
            'level': 'DEBUG',
            'propagate': True,
        },
    },
}
import os

# Absolute path to the directory where collectstatic will copy static files
STATIC_ROOT = os.path.join(BASE_DIR, 'staticfiles')

# Define the directories where static files are collected
STATIC_URL = '/static/'

# Where to find static files in the project
STATICFILES_DIRS = [
    os.path.join(BASE_DIR, 'static'),
]

import environ

# Initialize environment variables
env = environ.Env()
# Read the .env file if it exists
environ.Env.read_env()
