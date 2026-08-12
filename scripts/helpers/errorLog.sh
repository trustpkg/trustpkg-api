#!/bin/bash

set -e

max_width=65

red_color="\033[0;31m"
reset_color="\033[0m"

line=$(printf '%*s' "$max_width" '' | tr ' ' '-')

echo -e "${red_color}${line}${reset_color}"
echo -e "${red_color}[ ERROR ✗ ] $1 ${reset_color}" | fold -w $max_width
echo -e "${red_color}${line}${reset_color}"