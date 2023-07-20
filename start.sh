#!/bin/bash

output_file="./logs/app.txt"


while true; do
  # Start the command and redirect its output to a temporary file
  tmp_output_file=$(mktemp)
  run main.go > "$tmp_output_file" 2>&1

  # Store the exit status
  status=$?

  # Check the exit status
  if [ $status -ne 0 ]; then
    echo "The command exited with a non-zero status: $status"
  else
    echo "The command exited successfully."
  fi

  # Append the command logs to the output file
  cat "$tmp_output_file" >> "$output_file"
  rm "$tmp_output_file"

  # Display the command logs
  echo "Command logs:"
  tail -n +1 "$output_file"

  # Add a delay before running the command again
  sleep 1
done
