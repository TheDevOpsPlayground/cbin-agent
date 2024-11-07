from django.db import models

# Create your models here.
from django.db import models

class Agent(models.Model):
    ip_address = models.GenericIPAddressField(unique=True)
    status = models.BooleanField(default=False)  

    def __str__(self):
        return self.ip_address

    def get_status_display(self):
        return "Active" if self.status else "Inactive"
    

