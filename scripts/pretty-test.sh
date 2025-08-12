#!/usr/bin/env bash

jq -r '
  select(.Action == "fail" or .Action == "pass" or .Action == "run" or .Action == "output") |
  if .Action == "output" and (.Output != null) then
    # trim trailing newlines, preserve multiline content
    (.Output | sub("[\n]+$"; ""))
  elif .Action == "run" and (.Test != null) then
    "\n\u001b[33mEXECUTING \(.Test)\u001b[0m"
  elif .Action == "pass" and (.Test == null) then
    empty
  else
    "\(.Action) \(.Test)"
  end
'
