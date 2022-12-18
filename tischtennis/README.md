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

To sync s3 files:
```
serverless client deploy
```

To sync s3 files locally:
```
# Run minio on port 9000

# If bucket not created
AWS_ACCESS_KEY_ID=minioadmin AWS_SECRET_ACCESS_KEY=minioadmin aws s3api create-bucket --bucket tischtennis-local-static --region us-east-1 --endpoint http://localhost:9000
AWS_ACCESS_KEY_ID=minioadmin AWS_SECRET_ACCESS_KEY=minioadmin aws s3api put-bucket-policy --bucket tischtennis-local-static --policy file://policy.json --endpoint http://localhost:9000
# with policy.json:
##############
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "AllowPublicRead",
            "Effect": "Allow",
            "Principal": {
                "AWS": "*"
            },
            "Action": [
                "s3:GetObject"
            ],
            "Resource": [
                "arn:aws:s3:::tischtennis-local-static/*"
            ]
        }
    ]
}
###############

# Sync files
AWS_ACCESS_KEY_ID=minioadmin AWS_SECRET_ACCESS_KEY=minioadmin aws s3 cp ./static s3://tischtennis-local-static/static --recursive --endpoint http://localhost:9000


```
