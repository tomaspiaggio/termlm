function llm_generate_widget() {
  # Create a temporary file for the result
  local temp_file=$(mktemp)
  
  # Run termlm interactively, passing the temp file path
  termlm "$temp_file"
  local exit_code=$?
  
  # If successful and we got a result, put it in the command buffer
  if [[ $exit_code -eq 0 && -f "$temp_file" ]]; then
    local resp
    resp=$(cat "$temp_file")
    if [[ -n "$resp" ]]; then
      BUFFER="$resp"
      CURSOR=${#BUFFER}
      zle redisplay
    fi
  fi
  
  # Clean up
  rm -f "$temp_file"
}

zle -N llm_generate_widget
# bind to Ctrl‐K by default; override Cmd+K in your terminal to send Ctrl+K
bindkey '^K' llm_generate_widget
