#!/usr/bin/env bash

# job save
curl -i -XPOST http://0.0.0.0:1234/job -d 'name=job1&command=echo hello&cronExpr=*/5 * * * * * *'

# job list
curl -i -XGET http://0.0.0.0:1234

# job delete
curl -i -XDELETE http://0.0.0.0:1234/job/job1

# job kill
curl -i -XPUT http://0.0.0.0:1234/job/job1
