#!/bin/bash
# Script to analyze test coverage and identify uncovered code blocks

# Set colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Running tests with coverage...${NC}"
go test -v -coverprofile=coverage.out -covermode=atomic ./... || { 
  echo -e "${RED}Tests failed!${NC}" 
  exit 1
}

# Generate the coverage percentage
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
echo -e "${GREEN}Total coverage: ${COVERAGE}${NC}"

# Generate detailed report of uncovered functions
echo -e "\n${YELLOW}Functions with less than 100% coverage:${NC}"
go tool cover -func=coverage.out | grep -v "100.0%" | sort -k 3 -r

# Generate a list of uncovered code blocks with details
echo -e "\n${YELLOW}Generating detailed coverage report...${NC}"
echo -e "${YELLOW}Uncovered blocks of code:${NC}"

TMPFILE=$(mktemp)
go tool cover -html=coverage.out -o coverage.html

# Find files with uncovered lines
FILES=$(grep -l "cov0" coverage.html)

for FILE in $FILES; do
  # Extract source file path using sed instead of grep -P
  SOURCE_FILE=$(grep 'file="' "$FILE" | head -1 | sed 's/.*file="\([^"]*\)".*/\1/')
  if [ -z "$SOURCE_FILE" ]; then continue; fi
  
  echo -e "\n${GREEN}File: ${SOURCE_FILE}${NC}"
  
  # Extract uncovered line numbers with sed instead of grep -P
  UNCOVERED_LINES=$(grep -n 'cov0' "$FILE" | 
                  sed -E 's/.*id="L([0-9]+)".*class="cov0".*/\1/' |
                  awk '{print $1}')
  
  if [ -z "$UNCOVERED_LINES" ]; then continue; fi
  
  # Group consecutive line numbers
  echo "$UNCOVERED_LINES" | 
  awk 'BEGIN{prev=-2} 
  {
    if($1==prev+1){
      if(start==""){start=prev}
      end=$1
    } else {
      if(start!=""){
        print start"-"end
        start=""
        end=""
      } else if(prev!=-2){
        print prev
      }
      prev=$1
    }
  } 
  END{
    if(start!=""){
      print start"-"end
    } else if(prev!=-2){
      print prev
    }
  }' | while read -r LINE_RANGE; do
    
    # Print line range
    echo -e "${YELLOW}Lines $LINE_RANGE:${NC}"
    
    # Extract start and end line
    if [[ $LINE_RANGE == *-* ]]; then
      START=$(echo $LINE_RANGE | cut -d'-' -f1)
      END=$(echo $LINE_RANGE | cut -d'-' -f2)
    else
      START=$LINE_RANGE
      END=$LINE_RANGE
    fi
    
    # Print code block with context (3 lines before and after)
    CONTEXT=3
    START_WITH_CONTEXT=$((START-CONTEXT))
    END_WITH_CONTEXT=$((END+CONTEXT))
    
    # Ensure context doesn't go below line 1
    if [ $START_WITH_CONTEXT -lt 1 ]; then
      START_WITH_CONTEXT=1
    fi
    
    sed -n "${START_WITH_CONTEXT},${END_WITH_CONTEXT}p" "$SOURCE_FILE" | 
    awk -v start="$START" -v end="$END" -v context_start="$START_WITH_CONTEXT" '{
      line_num = context_start + NR - 1;
      if(line_num >= start && line_num <= end) {
        printf "  \033[31m%4d: %s\033[0m\n", line_num, $0;  # Red for uncovered lines
      } else {
        printf "  \033[90m%4d: %s\033[0m\n", line_num, $0;  # Gray for context
      }
    }'
  done
done

echo -e "\n${GREEN}HTML coverage report generated at: ${PWD}/coverage.html${NC}"
echo -e "${YELLOW}Open it with: open coverage.html${NC}"

rm -f $TMPFILE

# Provide a summary of where to focus testing efforts
echo -e "\n${YELLOW}Testing improvement focus areas:${NC}"
go tool cover -func=coverage.out | grep -v "100.0%" | sort -k 3 -n | head -5 