# Corporate Directory [![CircleCI](https://circleci.com/gh/tna0y/circleci-python-playing.svg?style=svg)](https://circleci.com/gh/tna0y/circleci-python-playing)

Solution to the test task

## Instructions
Bureaucr.at is a typical hierarchical organization. Claire, its CEO, has a hierarchy of employees reporting to her and each employee can have a list of other employees reporting to him/her. An employee with at least one report is called a Manager.
   
Your task is to implement a corporate directory for Bureaucr.at with an interface to find the closest common Manager (i.e. farthest from the CEO) between two employees. You may assume that all employees eventually report up to the CEO.
   
Here are some guidelines:
* Resolve ambiguity with assumptions.
* The directory should be an in-memory structure.
* A Manager should link to Employees and not the other way around.
* We prefer that you to use Go, but accept other languages too.
* How the program takes its input and produces its output is up to you.


## Assumptions
* All employees form a tree in their management relationships
* Root of the tree is indicated by an employee named "Claire"
* Tree is relatively static so we can afford to rebuild it for a full set of employees

## Algorithm solution
Taking into account all the assumptions task could be formulated as "Persistent tree online least common ancestor" problem.
I decided to implement DFS+RMQ solution, where RMQ is implemented via SQRT-decomposition. 

Solution provides us with `O(|V|)` preprocessing time complexity and `O(sqrt(|V|))` 
time complexity per query. Implementation details could be seen in comments to the code in
[pkg/lca/lca.go](pkg/lca/lca.go) file.

## Interface

Interface was implemented as a REST-like JSON rpc service. API specifications could be found in 
[swagger.yml](swagger.yml) file. I encourage you to load it into your favorite request making tool, e.g. Postman or open
 it on [swaggerhub](https://app.swaggerhub.com/apis/tna0y/CorporateDirectory/1.0.0).
 
All the backend stuff was implemented using go-kit and httprouter.

## DevOps

Dockerfile and docker-compose files are provided with the solution. Simple CI pipeline is also present, done with CircleCI.
Pipeline builds the docker image which includes all tests and submits the image to the Dockerhub.

Service is deployed on GCP GCE [http://35.234.96.95/employees](http://35.234.96.95/employees)

