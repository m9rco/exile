#!/usr/bin/env bash

# job save
curl -i -XPOST http://0.0.0.0:1234/job/save -d 'job={"name": "job1","command":"echo hello","cronExpr": "*/5 * * * * * *"}'

