#!/bin/bash

set -e

max_width=65

violet_color="\033[0;34m"
reset_color="\033[0m"

line=$(printf '%*s' "$max_width" '' | tr ' ' '-')

echo -e "${violet_color}${line}${reset_color}"
echo -e "${violet_color}[ RUN ⚙ ] $1 ${reset_color}" | fold -w $max_width
echo -e "${violet_color}${line}${reset_color}"