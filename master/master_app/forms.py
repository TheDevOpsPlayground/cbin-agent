from django import forms



from django import forms

class AddAgentForm(forms.Form):
    ip_address = forms.GenericIPAddressField(label='Agent IP Address', required=True)
