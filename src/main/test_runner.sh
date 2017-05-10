#!/bin/sh

tests=("echo hi" "echo bye" "echo thing")

session=$1
shift

tmux start-server
tmux new -d -s $session

on_error() {
  tmux kill-session -t $session
}
trap on_error ERR

shift
tmux send -t $session:0 "$cmd1" C-m

for i in tests
do
    tmux splitw -t $session -l 1
    tmux send -t $session:0.1 "$i" C-m
    tmux selectp -t $session:0.0
    tmux selectl -t $session tiled
done

# tmux setw synchronize-panes on
tmux a -t $session
