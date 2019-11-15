#!/bin/sh

# Start Kubernetes Route extender
./go/bin/kube-chorus-router -D &
status=$?
if [ $status -ne 0 ]; then
  echo "Failed to start Route Extender: $status"
  exit $status
fi

# Start the citrix node controller
./go/bin/citrix-node-controller -D &
status=$?
if [ $status -ne 0 ]; then
  echo "Failed to start citrix Node Controller: $status"
  exit $status
fi


while /bin/true; do
  ps aux |grep kube-chorus-router |grep -q -v grep
  PROCESS_1_STATUS=$?
  ps aux |grep citrix-node-controller |grep -q -v grep
  PROCESS_2_STATUS=$?
  # If the greps above find anything, they will exit with 0 status
  # If they are not both 0, then something is wrong
  if [ $PROCESS_1_STATUS -ne 0 -o $PROCESS_2_STATUS -ne 0 ]; then
    echo "One of the processes has already exited."
    exit -1
  fi
  sleep 60
done
