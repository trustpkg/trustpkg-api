#!/bin/bash

curl -s 'https://replicate.npmjs.com/registry/' | jq '.update_seq'