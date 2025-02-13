$schedule = new-object -com("Schedule.Service")
$schedule.connect() 
$tasks = $schedule.getfolder("\").gettasks(0)
$tasks  | Format-Table   Name , LastRunTime    # -AutoSize
IF($tasks.count -eq 0) {Write-Host “Schedule is Empty”}
Read-Host
