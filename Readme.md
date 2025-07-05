# AI-Bot

AI-Bot is a Slack bot powered by AWS Bedrock. It answers questions using a retrieval augmented approach and stores conversations in DynamoDB so it can keep track of each thread. The bot's responses are slightly sarcastic by design.
 
You can connect a Slack bot to this project and feed the datastore with the
information you need the model to train on.

This can help other teams directly get the info from the bot rather than
disturbing a real human.

## Setup

1. Configure AWS credentials and region. The service uses Bedrock and DynamoDB so the standard `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` and optional `AWS_REGION` variables must be available. The ID of your Bedrock knowledge base is provided through `KNOWELDGE_BASE_ID`.
2. Provide Slack credentials in `/etc/ai-chat-bot`:
   - `/etc/ai-chat-bot/slack-token` – Slack bot token
   - `/etc/ai-chat-bot/slack-signing-secret` – Slack signing secret
3. Run the application with Go:

```bash
go run ./...
```

You can also build the Docker image:

```bash
docker build -t ai-bot .
```

and run it with the appropriate environment variables and mounted credential files.

## What it does

The bot listens to Slack events, sends the message history to Bedrock and replies in the same thread. Threads are persisted in DynamoDB so it can continue conversations across messages.

- AWS keys injected as environment variables
- Slack workspace keys
