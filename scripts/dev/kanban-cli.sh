#!/bin/bash

API_URL="http://localhost:7070"

# Enums for Card
PRIORITIES=("High" "Medium" "Low")
STATUSES=("To Do" "In Progress" "In Review" "Done")
TYPES=("Bug" "Feature" "Maintenance")

# Helper: check if value is in array
in_array() {
  local val="$1"; shift
  for item; do [[ "$item" == "$val" ]] && return 0; done
  return 1
}

# Usage instructions
usage() {
  echo "Kanban CLI"
  echo "Usage:"
  echo "  $0 board list"
  echo "  $0 board get <boardId>"
  echo "  $0 board create <name> <description> <createdBy>"
  echo "  $0 board update <boardId> <name> <description>"
  echo "  $0 board delete <boardId>"
  echo "  $0 card list <boardId>"
  echo "  $0 card get <boardId> <cardId>"
  echo "  $0 card create <boardId> <title> <desc> <acceptanceCriteria> <status> <type> <priority> <storyPoints> <createdBy>"
  echo "  $0 card update <boardId> <cardId> <title> <desc> <acceptanceCriteria> <status> <type> <priority> <storyPoints>"
  echo "  $0 card delete <boardId> <cardId>"
  echo "  $0 dep list <boardId> <cardId>"
  echo "  $0 dep add <boardId> <cardId> <dependencyCardId>"
  echo "  $0 dep delete <boardId> <cardId> <dependencyCardId>"
  exit 1
}

# Board commands
board_list() { curl -s "$API_URL/boards" | jq; }
board_get() { curl -s "$API_URL/boards/$1" | jq; }
board_create() {
  if [[ ${#1} -lt 3 || ${#2} -lt 3 ]]; then
    echo "Board name and description must be at least 3 characters."; exit 1
  fi
  curl -s -X POST "$API_URL/boards" \
    -H "Content-Type: application/json" \
    -d "{\"boardName\":\"$1\",\"description\":\"$2\",\"createdBy\":\"$3\"}" | jq
}
board_update() {
  if [[ ${#2} -lt 3 || ${#3} -lt 3 ]]; then
    echo "Board name and description must be at least 3 characters."; exit 1
  fi
  curl -s -X PUT "$API_URL/boards/$1" \
    -H "Content-Type: application/json" \
    -d "{\"boardName\":\"$2\",\"description\":\"$3\"}" | jq
}
board_delete() { curl -s -X DELETE "$API_URL/boards/$1"; }

# Card commands
card_list() { curl -s "$API_URL/boards/$1/cards" | jq; }
card_get() { curl -s "$API_URL/boards/$1/cards/$2" | jq; }
card_create() {
  # Validate required fields
  if [[ ${#2} -lt 3 || ${#3} -lt 3 || ${#4} -lt 3 ]]; then
    echo "Title, description, and acceptanceCriteria must be at least 3 characters."; exit 1
  fi
  if ! in_array "$5" "${STATUSES[@]}"; then
    echo "Invalid status. Allowed: ${STATUSES[*]}"; exit 1
  fi
  if ! in_array "$6" "${TYPES[@]}"; then
    echo "Invalid type. Allowed: ${TYPES[*]}"; exit 1
  fi
  if ! in_array "$7" "${PRIORITIES[@]}"; then
    echo "Invalid priority. Allowed: ${PRIORITIES[*]}"; exit 1
  fi
  if ! [[ "$8" =~ ^[0-9]+$ ]]; then
    echo "storyPoints must be an integer."; exit 1
  fi
  curl -s -X POST "$API_URL/boards/$1/cards" \
    -H "Content-Type: application/json" \
    -d "{
      \"title\":\"$2\",
      \"description\":\"$3\",
      \"acceptanceCriteria\":\"$4\",
      \"status\":\"$5\",
      \"type\":\"$6\",
      \"priority\":\"$7\",
      \"storyPoints\":$8,
      \"createdBy\":\"$9\"
    }" | jq
}
card_update() {
  if [[ ${#3} -lt 3 || ${#4} -lt 3 || ${#5} -lt 3 ]]; then
    echo "Title, description, and acceptanceCriteria must be at least 3 characters."; exit 1
  fi
  if ! in_array "$6" "${STATUSES[@]}"; then
    echo "Invalid status. Allowed: ${STATUSES[*]}"; exit 1
  fi
  if ! in_array "$7" "${TYPES[@]}"; then
    echo "Invalid type. Allowed: ${TYPES[*]}"; exit 1
  fi
  if ! in_array "$8" "${PRIORITIES[@]}"; then
    echo "Invalid priority. Allowed: ${PRIORITIES[*]}"; exit 1
  fi
  if ! [[ "$9" =~ ^[0-9]+$ ]]; then
    echo "storyPoints must be an integer."; exit 1
  fi
  curl -s -X PUT "$API_URL/boards/$1/cards/$2" \
    -H "Content-Type: application/json" \
    -d "{
      \"title\":\"$3\",
      \"description\":\"$4\",
      \"acceptanceCriteria\":\"$5\",
      \"status\":\"$6\",
      \"type\":\"$7\",
      \"priority\":\"$8\",
      \"storyPoints\":$9
    }" | jq
}
card_delete() { curl -s -X DELETE "$API_URL/boards/$1/cards/$2"; }

# Dependency commands
dep_list() { curl -s "$API_URL/boards/$1/cards/$2/dependencies" | jq; }
dep_add() {
  curl -s -X POST "$API_URL/boards/$1/cards/$2/dependencies" \
    -H "Content-Type: application/json" \
    -d "[$3]" | jq
}
dep_delete() { curl -s -X DELETE "$API_URL/boards/$1/cards/$2/dependencies/$3"; }

# Main dispatcher
case "$1" in
  board)
    case "$2" in
      list) board_list ;;
      get) board_get "$3" ;;
      create) board_create "$3" "$4" "$5" ;;
      update) board_update "$3" "$4" "$5" ;;
      delete) board_delete "$3" ;;
      *) usage ;;
    esac
    ;;
  card)
    case "$2" in
      list) card_list "$3" ;;
      get) card_get "$3" "$4" ;;
      create) card_create "$3" "$4" "$5" "$6" "$7" "$8" "$9" "${10}" "${11}" ;;
      update) card_update "$3" "$4" "$5" "$6" "$7" "$8" "$9" "${10}" "${11}" ;;
      delete) card_delete "$3" "$4" ;;
      *) usage ;;
    esac
    ;;
  dep)
    case "$2" in
      list) dep_list "$3" "$4" ;;
      add) dep_add "$3" "$4" "$5" ;;
      delete) dep_delete "$3" "$4" "$5" ;;
      *) usage ;;
    esac
    ;;
  *) usage ;;
esac