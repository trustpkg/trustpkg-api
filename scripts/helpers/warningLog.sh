#!/bin/bash

set -e

max_width=65

orange_color="\033[0;33m"
reset_color="\033[0m"

line=$(printf '%*s' "$max_width" '' | tr ' ' '-')

echo -e "${orange_color}${line}${reset_color}"
echo -e "${orange_color}[ WARNING ⚠ ] $1 ${reset_color}" | fold -w $max_width
echo -e "${orange_color}${line}${reset_color}"