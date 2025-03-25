#!/bin/bash

COUNTER_FILE=".script_counter"

if [ ! -f "$COUNTER_FILE" ]; then
    echo "1" > "$COUNTER_FILE"
    COUNTER=1
else
    COUNTER=$(cat "$COUNTER_FILE")
    COUNTER=$((COUNTER + 1))
    echo "$COUNTER" > "$COUNTER_FILE"
fi

ANSWER_FILE="answer_${COUNTER}.txt"

case $COUNTER in
    1)
        echo "Question 1: What is 10 + 15?"
        echo "Please write your answer in a file named '$ANSWER_FILE' and then run ./my-script.sh again"
        ;;
    2)
        echo "Question 2: What is the capital of France?"
        echo "Please write your answer in a file named '$ANSWER_FILE' and then run ./my-script.sh again"
        ;;
    3)
        echo "Question 3: What is the chemical symbol for water?"
        echo "Please write your answer in a file named '$ANSWER_FILE' and then run ./my-script.sh again"
        ;;
    4)
        echo "Question 4: Who wrote 'Romeo and Juliet'?"
        echo "Please write your answer in a file named '$ANSWER_FILE' and then run ./my-script.sh again"
        ;;
    5)
        echo "Question 5: What is the square root of 16?"
        echo "Please write your answer in a file named '$ANSWER_FILE' and then run ./my-script.sh again"
        ;;
    6)
        echo "Checking your answers..."
        
        # Check if all answer files exist
        for i in {1..5}; do
            if [ ! -f "answer_${i}.txt" ]; then
                echo "Missing answer for question $i. Please complete all questions."
                exit 1
            fi
        done
        
        echo "Good job! You've answered all the questions!"
        
        # Clean up
        rm -f "$COUNTER_FILE"
        echo "Do you want to clean up all answer files? (y/n)"
        read -r response
        if [ "$response" = "y" ]; then
            rm -f answer_*.txt
            echo "All files cleaned up. Thanks for playing!"
        else
            echo "Files kept for your reference. Thanks for playing!"
        fi
        ;;
    *)
        echo "All questions completed!"
        # Clean up
        rm -f "$COUNTER_FILE"
        ;;
esac
