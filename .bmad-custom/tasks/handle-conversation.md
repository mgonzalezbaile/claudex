# Handle Conversation Task

This task provides instructions for executing the Voiced simulator to manage conversation sessions. The simulator handles session initialization automatically through the backend, so the agent's primary responsibility is to execute the simulator correctly in single-shot mode.

## Purpose

- Execute the Voiced simulator in single-shot mode for conversation testing
- Start new conversations and continue existing sessions using session IDs
- Maintain proper persona adoption throughout conversation flows
- Support both text messages and reflection card sessions

## Simulator Overview

The Voiced simulator is located at `/Users/maikel/Workspace/Pelago/voiced/simulator` and provides:
- **Single-shot mode**: Execute one conversation exchange and return detailed analysis
- **Session management**: Automatic session ID generation or continuation with provided session ID
- **Reflection support**: Send reflection cards (intro or normal) to trigger onboarding instructions
- **User creation**: Create users with custom data for testing specific scenarios
- **Authentication**: Automatic auth token handling with Firebase emulator

## Key Simulator Commands

### Creating Custom Users

**Create User with Specific Data:**
```bash
cd /Users/maikel/Workspace/Pelago/voiced/simulator
yarn dev --create-user='{"userId":"custom-user-123","username":"testuser","onboarded":true,"totalSessions":5}'
```

### Starting New Conversations

**Text Message with Specific User ID:**
```bash
cd /Users/maikel/Workspace/Pelago/voiced/simulator
yarn dev --one-shot --user-id="advanced-user-123" --message="Your persona message here"
```

**Reflection Session with Specific User ID:**
```bash
cd /Users/maikel/Workspace/Pelago/voiced/simulator
yarn dev --one-shot --user-id="advanced-user-123" --reflection-intro --message="Your persona message here"
```

### Continuing Existing Conversations

**Continue with Session ID and Specific User ID:**
```bash
cd /Users/maikel/Workspace/Pelago/voiced/simulator
yarn dev --one-shot --session-id="sim_1234567890_abc123def" --user-id="advanced-user-123" --message="Your follow-up message"
```

## Critical Logging Requirements

**MANDATORY WORKFLOW FOR ALL CONVERSATIONS:**
1. Send message via simulator
2. Receive response from Voiced
3. IMMEDIATELY update log file with both message and response
4. Only then proceed to next message

**NEVER batch log updates** - each exchange must be logged before continuing.