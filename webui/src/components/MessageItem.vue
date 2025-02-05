<!-- File: webui/src/components/MessageItem.vue -->
<template>
    <div class="mb-2 message-item">
      <div class="message-header">
        <strong>{{ message.senderName || 'Unknown' }}</strong>
        <span class="separator"></span>
        <small class="text-muted">{{ formatDate(message.timestamp) }}</small>
      </div>
      <div class="message-content">{{ message.content }}</div>
      <div class="message-actions mt-1">
        <!-- Reaction Button -->
        <button class="btn btn-sm btn-outline-secondary me-1" @click="$emit('react', message)">React</button>
        <!-- Forward Button -->
        <button class="btn btn-sm btn-outline-secondary me-1" @click="$emit('forward', message)">Forward</button>
        <!-- Comment Button -->
        <button class="btn btn-sm btn-outline-secondary" @click="$emit('comment', message)">Comment</button>
      </div>
      <!-- Display reactions if any -->
      <div class="message-reactions mt-1" v-if="message.reactions && message.reactions.length">
        <span v-for="(reaction, index) in message.reactions" :key="index" class="reaction">
          {{ reaction.emoji }} ({{ reaction.count }})
        </span>
      </div>
      <!-- Display comments if any -->
      <div class="message-comments mt-2" v-if="message.comments && message.comments.length">
        <div v-for="comment in message.comments" :key="comment.commentId">
          <CommentItem :comment="comment" @delete="($event) => $emit('deleteComment', message.id, $event)" />
        </div>
      </div>
    </div>
  </template>
  
  <script setup>
  import { format } from 'date-fns'
  import CommentItem from './CommentItem.vue'
  
  const props = defineProps({
    message: {
      type: Object,
      required: true
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
  </script>
  
  <style scoped>
  .separator {
    margin: 0 4px;
    font-weight: normal;
  }
  .message-item {
    border-bottom: 1px solid #e0e0e0;
    padding-bottom: 8px;
  }
  .message-actions button {
    font-size: 0.8rem;
  }
  .reaction {
    margin-right: 6px;
    font-size: 1.1rem;
  }
  </style>
  