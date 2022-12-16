# Tischtennis
Tischtennis on serverless

Binary static files are hosted on google drive and the link is directly placed into the html since I wasn't able to get binary static files to work

env.dev.json
```
{
  "MASTER_ACCESS_KEY": "...",
  "BASE_PATH": "dev",
  "VERSION": "0.1",
  "ADD_FAKE_DATA": "true"
}
```


To run offline:
```
sudo ip addr show docker0
```
Get inet IP address from the above command, put into the code/database/database.go config for connecting to ddb

Now run
```
sls offline start
```
