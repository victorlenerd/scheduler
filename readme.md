# Scheduler0

A simple scheduling server for apps and backend server.

[![Go Report Card](https://goreportcard.com/badge/github.com/victorlenerd/scheduler0)](https://goreportcard.com/report/github.com/victorlenerd/scheduler0) 
[![CircleCI](https://circleci.com/gh/victorlenerd/scheduler0/tree/master.svg?style=svg)](https://circleci.com/gh/victorlenerd/scheduler0/tree/master)

## Current Status

To use this you must first clone the repository and add the scheduler0 binary at the root of repo to your $PATH
    
The scheduler0 CLI tool in the root of this repo can be used to configure and start the server.
Postgres is needed to start the server, therefore after cloning the repo you need to run the config command.
```shell
scheduler0 config init
```

Will take you through the configuration flow, where you will be prompted to enter your database credentials.
To start the http server.
```shell
scheduler0 start
```

For more information. Use the help flag
```shell
scheduler0 --help
```

## API Documentation

There is a REST API documentation on http://localhost:9090/api-docs/ [](http://localhost:9090/api-docs/)

Note: that port 9090 is the default port for the server.

## Example Usage In Node Server

```javascript
'use strict'

/**
 * The purpose of this program is to test the scheduler0 server.
 * It sends a request scheduler0 server and counts scheduler0 request to the callback url.
 * It's kinda like a scratch pad.
 * **/

require('dotenv').config()
const axios = require('axios')
const express = require('express')

const app = express()
const port = 3000

// Scheduler0 environment variables
const scheduler0Endpoint = process.env.API_ENDPOINT
const scheduler0ApiKey = process.env.API_KEY
const scheduler0ApiSecret = process.env.API_SECRET

const axiosInstance = axios.create({
    baseURL: scheduler0Endpoint,
    headers: {
        'x-api-key': scheduler0ApiKey,
        'x-secret-key': scheduler0ApiSecret
    }
});

async function createProject() {
    const { data: { data } } = await axiosInstance
        .post('/projects', {
            name: "sample project",
            description: "my calendar project"
        });
    return data
}

async function createJob(projectUUID) {
    try {
        const { data: { data } } = await axiosInstance
            .post('/jobs', {
                name: "sample project",
                spec: "@every 1m",
                project_uuid: projectUUID,
                callback_url: "http://localhost:3000/callback"
            });

        console.log(data)

        return data
    } catch (err) {
        console.log({ error: err.response.data })
    }
}

// This callback will get executed every minute
app.post('/callback', (req, res) => {
    res.send(`Callback executed at :${(new Date()).toUTCString()}`)
})

app.listen(port, async () => {
    const project = await createProject()
    const job  = await createJob(project.uuid)
   
    console.log(`app listening at http://localhost:${port}`)
})

```


## [WIP]: Dashboard

The dashboard is supposed to be a GUI for managing projects, credentials and jobs.

!["Dashboard"](./screenshots/screenshot.png)
 
# License

 * MIT license ([LICENSE-MIT](LICENSE-MIT) or
   http://opensource.org/licenses/MIT)

at your option.

### Contribution

Unless you explicitly state otherwise, any contribution intentionally submitted
for inclusion in this project by you, as defined in the Apache-2.0 license,
shall be dual licensed as above, without any additional terms or conditions.
