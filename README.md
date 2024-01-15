# Scalable Webhook Delivery System:

### Objective:
To develop an efficient, reliable system capable of:

- High Volume Delivery
- Continuously monitoring a specific folder for new files.
- Processing each file quickly and accurately to extract user IDs.
- Handling the high volume of webhook calls while respecting any rate limits.
- Ensuring data integrity and delivery confirmation for each alert message.

For more context or problem statement refer: problem_statement.md

### About:

This whole repo has 3 main services and redis

#### Alert Initiator:  

- This service monitors the  `~/.tmp` in the local system which is mounted to `/tmp` file of the container
- Also this monitors the queues, db (In this code, for simplicity a key in redis has been used instead of any RDMS)
- This service also has a httpserver for debugging and ease of use

#### Alert Worker:

-  This service has the workers which constantly polls the queues (batcher queue and notifier queue) for the tasks to execute them.

#### Alert Notifier:

-  This the webhook server which receives the webhooks and notifies the respective users with the alert message. (Here it just appends the user_id,alert_message to the file inside `~/.tmp/output`)
- This server has ratelimiter implemented. ***To configure the ratelimiter rps, change RatelimiterRPS in `constants/server.go`file*** 
> Note: Currently the RPS is .5 which means 1 request per 2 seconds (which is too slow, this is present for testing all failcases ;) ) is accepted for the constraints (it's been added as ip ratelimiter so the constraints on ip and requested path). If it is changed to 5 then it means 5 request per second.


### Instructions:

#### Deployment

- Deploy redis first by `make local`
- Then run `make all` which deploys all the 3 services
- To deploy each service separately, you can use 
    ```
    alert_initator -> make server
    alert_worker -> make worker
    alert_notifier -> make notifier
    ```

#### Testing

- The test scripts will be present under `scripts`
    - `ratelimiter.sh` -> continues hits 10 times(you can change it) the alert_notifier for testing the ratelimiter
    - `script.sh` -> takes the number of rows as an argument and generates
    - After generating the file, copy it to `~/.tmp` in your system or you can use `/upload` API (More Info about this will be explained under APIS section) and check the contents of file under `~/.tmp/output`
    - To test data integrity, what happens if worker fails you can test it by making edits as mentioned in the code snippet comments end of  `alert_intiator/internal/services/worker/worker.go` func `ProcessBatcherTopic` 
    ```
    // test for data integrity as we shouldn't miss any user nor alerts should be sent to the user multiple times
	// if offset  == 2{
	// 	return fmt.Errorf("exiting")
	// }
    ```
    Uncomment the above snippet, the worker exists after reading 2 lines, we can recipocrate the situation of worker ending abruptly by making this change

#### APIS

- There are 3 APIS which are for debugging (there isn't any check on method)
    - `/upload` -> post api,accepts form-data body with file as key name, to upload the file
    - `/retry` ->  get api, to retry the rejected tasks in both the queues
    - `/webhook` -> post api, send a webhook to the alert_notifier, please find the sample body 
    ```
        {
            "userId": "125",
            "alertMessage": "Order has been placed"
        }
    ```

**For approach and design please check: Approach Doc and design.png**

#### Improvements:

- [ ] Add TestCases
- [ ] Add conf
- [ ] Remove info logs which are debug
- [ ] Add Context down the function calls which helps in profiling
- [ ] Add Routers,Middlewares, recovering in controllers
- [ ] Improve Directory Structure (As this is poc, didn't concentrated on this part)
- [ ] Along with using db/persistent storage, incase of backpressure try to slow down the injested rate in batcher queue
- [ ] Add more info like each service of alert_initator, worker etc