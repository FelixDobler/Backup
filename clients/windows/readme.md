- Use WSL Version 1 for useable performance on host os file systems
- copy and adjust [backup.ps1](windows_client/backup.ps1)
- change ownership to `Administrators` and remove write permission for normal users
  - this prevents manipulation of the while which will later be called with administrator privileges
- Create task in `Task Scheduler` from template
- Install Powershell module [BurntToast](https://github.com/Windos/BurntToast)
