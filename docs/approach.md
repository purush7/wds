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
