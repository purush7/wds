# Scalable Webhook Delivery System:

### Objective:
To develop an efficient, reliable system capable of:

- High Volume Delivery
- Continuously monitoring a specific folder for new files.
- Processing each file quickly and accurately to extract user IDs.
- Handling the high volume of webhook calls while respecting any rate limits.
- Ensuring data integrity and delivery confirmation for each alert message.

For more context or problem statement refer: 



### Instructions:






### Todo:

- [x] Create a global variable which has info of id,path,count
- [x] Start Consumers and its functions
- [x] Worker
- [x] Publish functions
- [x] AlertHandler
- [x] Queue size, what happens if task count is more than queue size?
- [x] what happens if rejected task count is more than queue size? and what is rejected task capacity/DLQs?
- [x] add mode to webhook server which takes ratelimiting threshold, mode which gives 500 always for testing, add flag for intiator-server rate limiting
- [x] Constraints on file uploaded size, alert message string length (Remove validation on csv file and ratelimiting to support high traffic and large input)
- [x] Clear the directory structure and use constants
- [x] Add script to start redis,worker,alert_system, webhook server containers
- [ ] Handle Back Pressure of queue and give 429
- [ ] Add script which generates huge csv file
- [ ] Error checks
    - [ ] check whether streaming/seeking is working properly or not
    - [ ] check large files 
    - [ ] retry and hanle the cases where webhook server fails,down
    - [ ] Testcases and testfiles
        - [ ] ratelimiting of webhook server <<<< file row size
        - [ ] huge files like 1-2GB and call 10-100 times of this file for alerting
        - [ ]

- [ ] Doc
    - [ ] Continuously monitoring a specific folder for new files. (S3 event -> lambda, locally trigger the api when push happens)
    - [ ] Processing each file quickly and accurately to extract user IDs.
    - [ ] Handling the high volume of webhook calls while respecting any rate limits.
    - [ ] Ensuring data integrity and delivery confirmation for each alert message.
    - [ ] Clear Instructions for deployment
    - [ ] How to test all error scenarios
    - [ ] Designs, modular and factors for each component
    - [ ] API documentation (swagger spec)
    - [ ] mention status api
    - [ ] Improvements



### Improvements:

- Directory structure 
    As this is a small POC and also assignment. I didn't structured it properly(limited time), tried to use less files/packages. This can be splitted into have request handlers,controllers, services, models etc.
- Pass the request context down the line, it will be useful for profiling, can have basic request info and  better handling incase of context cancel scenarios
- Add Routers,Middlewares, recovering in controllers



## context/description:


## Design: and other possible designs

- Large files,
- Many of requests (Rate limiting, backpressure)



## Project Directory:

- each service,file and its purpose