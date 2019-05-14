#!/bin/bash

curl --data "password=angryMonkey" http://localhost:8080/hash &
curl --data "password=crazyApe" http://localhost:8080/hash &
curl --data "password=tarzan123" http://localhost:8080/hash &
curl --data "password=jungleRadio" http://localhost:8080/hash &
curl --data "password=kinglouie" http://localhost:8080/hash &
curl --data "password=bananaz" http://localhost:8080/hash &
curl --data "password=panther" http://localhost:8080/hash &
curl --data "password=leopard" http://localhost:8080/hash &

loop="id not found"
while [ "$loop" = "id not found" ]
do
  echo "Getting hashed password for password=kinglouie"
  loop="$(curl -s http://localhost:8080/hash/5)"
  echo "Loop: $loop"
done


curl -i http://localhost:8080/stats

curl --data "password=angryMonkey" http://localhost:8080/hash &
curl --data "password=crazyApe" http://localhost:8080/hash &
curl --data "password=tarzan123" http://localhost:8080/hash &
curl --data "password=jungleRadio" http://localhost:8080/hash &
curl --data "password=kinglouie" http://localhost:8080/hash &
curl --data "password=bananaz" http://localhost:8080/hash &
curl --data "password=panther" http://localhost:8080/hash &
curl --data "password=leopard" http://localhost:8080/hash &
sleep 6

curl -i http://localhost:8080/stats

sleep 1
curl -i http://localhost:8080/shutdown


# below should fail to connect

curl --data "password=angryMonkey" http://localhost:8080/hash &
curl --data "password=crazyApe" http://localhost:8080/hash &
curl --data "password=tarzan123" http://localhost:8080/hash &
curl --data "password=jungleRadio" http://localhost:8080/hash &
curl --data "password=kinglouie" http://localhost:8080/hash &
curl --data "password=bananaz" http://localhost:8080/hash &
curl --data "password=panther" http://localhost:8080/hash &
curl --data "password=leopard" http://localhost:8080/hash &

curl -i http://localhost:8080/stats &
