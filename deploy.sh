#!/bin/bash

DST=til-main.tailaf38ca.ts.net:/opt/tex-to-pdfa/

rsync -avz ./ "$DST"
