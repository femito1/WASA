<!-- File: webui/src/components/MessageItem.vue -->
<template>
  <div class="mb-2 message-item">
    <!-- Reply reference for messages (if any) -->
    <div v-if="message.replyTo" class="reply-reference">
      <div class="reply-info">
        <span class="reply-sender">{{ message.replyTo.senderName }}</span>
        <span class="reply-snippet">{{ truncateContent(message.replyTo.content) }}</span>
      </div>
    </div>

    <!-- Display forwarded indicator if the message is forwarded -->
    <div v-if="message.isForwarded" class="forwarded-label">Forwarded</div>

    <div class="message-header">
      <strong>{{ message.senderName || 'Unknown' }}</strong>
      <span class="separator">•</span>
      <small class="text-muted">{{ formatDate(message.timestamp) }}</small>
      <span class="message-status ms-2">
        {{ message.state === 'Read' ? '✓✓' : '✓' }}
      </span>
    </div>

    <!-- Message content with image support -->
    <div class="message-content">
      <img v-if="message.format === 'image'" :src="message.content" 
           class="img-fluid message-image" alt="Shared image" />
      <span v-else>{{ message.content }}</span>
    </div>

    <div class="message-actions mt-1">
      <button class="btn btn-sm btn-outline-secondary me-1" @click="$emit('react', message)">
        React
      </button>
      <button class="btn btn-sm btn-outline-secondary me-1" @click="$emit('reply', message)">
        Reply
      </button>
      <button class="btn btn-sm btn-outline-secondary me-1" @click="$emit('forward', message)">
        Forward
      </button>
      <button
        v-if="currentUserId === message.senderId"
        class="btn btn-sm btn-outline-danger"
        @click="$emit('deleteMessage', message)"
      >
        Delete
      </button>
    </div>

    <!-- Display reactions -->
    <div class="message-reactions mt-1" v-if="message.reactions && message.reactions.length">
      <span v-for="reaction in message.reactions" :key="reaction.emoji" 
            class="reaction" @click="$emit('removeReaction', message, reaction.emoji)">
        {{ reaction.emoji }} {{ reaction.count }}
      </span>
    </div>
  </div>
</template>

<script setup>
import { format } from 'date-fns'

const props = defineProps({
  message: {
    type: Object,
    required: true,
  },
  currentUserId: {
    type: Number,
    required: true,
  }
})

function formatDate(dateString) {
  if (!dateString) return ''
  try {
    return format(new Date(dateString), 'p')
  } catch (e) {
    return dateString
  }
}

function truncateContent(content) {
  if (!content) return ''
  return content.length > 50 ? content.substring(0, 47) + '...' : content
}
</script>

<style scoped>
.message-item {
  border-bottom: 1px solid #e0e0e0;
  padding-bottom: 8px;
}

.message-image {
  max-width: 300px;
  border-radius: 4px;
}

.reaction {
  margin-right: 6px;
  padding: 2px 6px;
  background: #f0f0f0;
  border-radius: 12px;
  cursor: pointer;
  user-select: none;
}

/* New styles for WhatsApp-like reply reference */
.reply-reference {
  display: flex;
  align-items: center;
  border-left: 3px solid #25D366;
  padding-left: 8px;
  margin-bottom: 4px;
  background: #e9f7f1;
  border-radius: 4px;
}

.reply-info {
  display: flex;
  flex-direction: column;
}

.reply-sender {
  font-weight: bold;
  color: #075e54;
  font-size: 0.9em;
}

.reply-snippet {
  color: #333;
  font-size: 0.85em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Style for forwarded messages */
.forwarded-label {
  font-size: 0.8em;
  color: #555;
  opacity: 0.8;
  margin-bottom: 4px;
  font-style: italic;
}

.message-header .separator {
  margin: 0 4px;
}

.message-status {
  color: #28a745;
  font-size: 0.8em;
}
</style>
