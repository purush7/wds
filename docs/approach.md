## Approach

#### Objective:
To develop an efficient, reliable system capable of:

- Continuously monitoring a specific folder for new files.
- Processing each file quickly and accurately to extract user IDs.
- Handling the high volume of webhook calls while respecting any rate limits.
- Ensuring data integrity and delivery confirmation for each alert message.
- No multiple alerts for an user
- High Volume Delivery
- Horizontal Scalable


There are many ways to approach [this]((https://github.com/purush7/wds/blob/main/docs/problem_statement.md)) problem statement. Let's start from basic solution

##### Basic Solution:

- One service which monitors the folder 
- Read the new files once there is a create file event
- Send the webhook for that user with the respective message

This solution fails for the above objectives:

 - **High Volume Delivery** *(Missing of user alerts)*:
    ```In this case, it may start reading the file even though the file is getting uploaded. Let suppose there are 10+ users but this service can stop at 10 users.```

 - **Horizontal Scalable** *(Multiple Alerts)*:
    ```In this case, both the ec2/servers can read the same file and can send the multiple alerts to an user.```

 - **Handling Rate LImiters** *(Quick File Processing)*: 
     ```In this case, if there are 429/too many requests, the solutions is to wait and retry this will hinder the file processing or if suppose we have those in separate go-routines, then this may create multiple go-routines. To solve that if we have go routine pool, which will make file processing slow```

 - **FailSafe** *(Missing of user alerts)*:
    ```If we follow above solution of spawning many go routines or alteast using only one thread and during deployments, this entire thing will get and whole process isn't running in transaction format to rollback. So we will lose user alerts```

##### Conclusions:

- From above solution, there are 2 components file processing and sending webhooks.
- Both should be separate as one shouldn't affected other drastically
- Handling of 429s
- After every processing and movement of data from 1 component to other, there should be a tracking (can be on each row or overall as reading file is in sync) in persistent storage
- Each Component should be horizontal scalable irrespective of other

##### Analysis:

- To solve, first and last the process should be in async and we have to use **MQ(message queue) and workers**
- For every read of row, we need to track that it has been read and pushed
- As file processing is in sync, it will be enough to just track the row Count after pushing to the queue
- To handle 429s, we should add a wait/sleep incase response is 429 and retry. But if the webhook server rate limiters has been configured (decreased the accepted count) while delivery system is running, then even retry won't work. So it will go **rejected/dead-letter queue**
- As incase of 429s, we need to decrease the injection speed or in other words, there should be way of **back pressure** to the producer incase of 429s as the speed of ejection will be slower.
- Most of the cases file read will be faster (even incase of s3). To improve we can stream the file from s3 or even we can do **seek and read** (Here golang csvreader package doesn't support seek, but we can do it through normal file read and split string with `,`)
- The main blocker is the 429, if suppose it accepts larger rps then file processing can be the blocker, then we can just horizontal scale the workers
- To support **High Volume** and having **back pressure**, there are 2 ways one is to push into queue continuosly even tho there is back pressure as all the tasks will stay in queue but this makes redis slower. The other better way is to store the ready messages in the database(here I am using redis key for simplicity/poc)
- Monitor system should monitor the `folder`,`DLQ/rejected queues` and `database entries`
- Here to handle the back-pressure, I am setting a key `tooManyRequests`. If the key set, monitor won't push the rejected tasks and tasks stored in database to ready queue. If it isn't then it will push at regular intervals
- **High Volume File Processing**, the idea behind to solve this issue is whenever there is a event for create file, then check whether this file has been opened in other programs for writing. Currently this is being done checking for file flags. You can find more info in `isFileOpen`. (This procedure is supported in linux/unix anywhere this process is in linux container). 
- This won't be a problem incase of s3 folders as we can get alert after completion of uploading a file in s3 and which can trigger lambda(monitor service)
- As there are 2 components, here we need 2 queues 1 for file processing, and second for sending webhooks

#### Summary:

- Continuously monitoring a specific folder for new files. (Monitor service)
- Processing each file quickly and accurately to extract user IDs. (async process)
- Handling the high volume of webhook calls while respecting any rate limits. (DLQ and push into ready queue at intervals)
- Ensuring data integrity and delivery confirmation for each alert message. (Track the row count)
- No multiple alerts for an user (DLQ)
- High Volume Delivery (Monitor service,check file flags)
- Horizontal Scalable (MQ,Workers)

#### Final Solution:

You can find the design [here](https://github.com/purush7/wds/blob/main/docs/design.jpg).

- Monitor service:
    - monitors the folder and pushes the filepath,id into q1
    - monitor the queues and inject the DLQ tasks into ready queue at regular intervals if there isn't 429(A redis key)
    - monitor the db, and push the tasks stored in db to queues

- Workers:
    - Q1(batcher queue) which gets the filepath,id and process the file and push user_id,alerts into 2nd queue
    - Store the rowCount in a redis key after the reading rows for an id
    - Incase of backpressure(429) store the tasks into db as it may lead to overflowing of tasks and slows down the redis
    - Q2(notifier queue) send webhook, incase of 429 wait and set the redis key for backpressure. Incase of many retry fails, send the task to DLQ which monitor pushes again

- Redis:
    -  Store the data of queues and info
    - backpressure key
    - And ID with row count
    - A dbstore which acts like db to store the tasks incase of backpressure (for avoiding overflooding of tasks in the queue)
## Approach

### Objective:

Develop an efficient, reliable system capable of:

- Continuously monitoring a specific folder for new files.
- Processing each file quickly and accurately to extract user IDs.
- Handling the high volume of webhook calls while respecting any rate limits.
- Ensuring data integrity and delivery confirmation for each alert message.
- No multiple alerts for a user.
- High Volume Delivery.
- Horizontal Scalability.

### Basic Solution:

- One service monitors the folder.
- Reads new files once a create file event occurs.
- Sends the webhook for that user with the respective message.

### Challenges with Basic Solution:

- **High Volume Delivery (Missing User Alerts):**
  - May start reading the file before it finishes uploading, potentially missing user alerts.

- **Horizontal Scalability (Multiple Alerts):**
  - Multiple instances can read the same file, leading to multiple alerts for a user.

- **Handling Rate Limiters (Quick File Processing):**
  - Slow file processing due to rate limit issues.

- **FailSafe (Missing User Alerts):**
  - In case of deployments or errors, there is a risk of losing user alerts.

### Conclusions:

- Separate components for file processing and sending webhooks.
- Tracking for every processed row to avoid duplication.
- Asynchronous processing with MQ (Message Queue) and workers.
- Handling of 429s and Dead-Letter Queue for rejected tasks.
- Back pressure implementation to handle rate limits.
- Streaming from S3 or using seek and read for faster file processing.
- Each Component should be horizontal scalable irrespective of other

### Analysis:

- Asynchronous processing using MQ and workers for scalability.
- Tracking row count after pushing to the queue.
- Handling 429s with wait/retry and Dead-Letter Queue.
- Back pressure to manage injection speed and avoid overflow.
- Streaming or seek and read for high-volume file processing.
- 2 components: file processing and sending webhooks.
- Use of a database to store tasks during back pressure.
- Monitor system to track the folder, DLQ/rejected queues, and database entries.
- To solve issue of larger files, check whether the file has been opened in other programs for writing by using file flags. You can find more info in `isFileOpen`. 
> Note: This procedure is supported in linux/unix anywhere this process is in linux container. 

### Summary:

- Continuous folder monitoring (Monitor service).
- Quick and accurate file processing (Async process).
- Webhook delivery with rate limit handling (DLQ and periodic injections).
- Ensuring data integrity and no multiple alerts for a user (DLQ and track the row count).
- High Volume Delivery (Monitor service, check file flags).
- Horizontal Scalability (MQ, Workers).

### Final Solution:

[Design](https://github.com/purush7/wds/blob/main/docs/design.jpg)

#### Monitor Service:
- Monitors the folder, pushes filepath and ID into Q1.
- Monitors queues, injects DLQ tasks into the ready queue at intervals if no 429 (controlled by a Redis key).
- Monitors the database, pushes tasks stored in the database to queues.

#### Workers:
- Q1 (Batcher Queue): Processes file, pushes user_id and alerts into Q2.
- Stores rowCount in a Redis key after reading rows for an ID.
- In case of back pressure (429), stores tasks in the database.
- Q2 (Notifier Queue): Sends webhook, handles back pressure, and retries.
- Sends tasks to DLQ if retries fail.

#### Redis:
- Stores queue data and information.
- Manages backpressure with a key.
- Tracks IDs with row counts.
- Acts as a store for tasks during back pressure.

#### Notifier (Webhook-Server):
- Replicates the webhook server.
- Takes user_id and alert message.
- Enforces rate limiting.

- Notifier(webhook-server):
    - To replicate the webhook server, which takes the user_id,alert message
    - Has Rate limiter