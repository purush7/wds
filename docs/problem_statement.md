### Problem Statement:

You have been tasked with designing and developing a high-volume webhook delivery service.

The goal is to create a system that can process large files containing user IDs and send alert
messages to those users via Swilly's backend webhook API.

### File Processing: 
- The support team at Swilly will upload files containing between 1,000
to 100,000 lines to a designated folder. Each line in the file represents a single user with their unique userId.

### Webhook API Integration: 
- The system must make API calls to Swilly's webhook API, which accepts two parameters: userId and alertMessage. The alert message to be sent
will be provided.
- The files are in a standard format (e.g., CSV, JSON) with a clear structure. 
- The webhook API is capable of handling high throughput and has rate-limiting constraints, if any. 
- The system should be scalable to handle potential increases in file sizes or frequency. Appropriate error handling is required for failed deliveries or corrupted data entries. 
- The system must ensure
no user is missed and alerts are not sent multiple times to the same user.

### Objective:
To develop an efficient, reliable system capable of:

- High Volume Delivery
- Continuously monitoring a specific folder for new files.
- Processing each file quickly and accurately to extract user IDs.
- Handling the high volume of webhook calls while respecting any rate limits.
- Ensuring data integrity and delivery confirmation for each alert message.
