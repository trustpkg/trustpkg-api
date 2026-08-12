#!/bin/bash

set -e

max_width=65

green_color="\033[38;5;34m"
reset_color="\033[0m"

line=$(printf '%*s' "$max_width" '' | tr ' ' '-')

echo -e "${green_color}${line}${reset_color}"
echo -e "${green_color}[ SUCCESS ✓ ] $1 ${reset_color}" | fold -w $max_width
echo -e "${green_color}${line}${reset_color}"