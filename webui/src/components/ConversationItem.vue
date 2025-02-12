<!-- File: webui/src/components/ConversationItem.vue -->
<template>
    <li class="list-group-item list-group-item-action" @click="$emit('open', conversation)">
      <div class="d-flex justify-content-between align-items-center">
        <div class="d-flex align-items-center">
          <img v-if="conversation.picture" :src="conversation.picture" alt="Conversation Photo" class="conversation-photo me-2" />
          <strong>{{ conversation.name || 'Unnamed Conversation' }}</strong>
        </div>
        <small class="text-muted">{{ formatDate(conversation.updatedAt) }}</small>
      </div>
    </li>
  </template>
  
  <script setup>
  import { toDisplayString } from 'vue'
  import { format } from 'date-fns'
  
  defineProps({
    conversation: {
      type: Object,
      required: true
    }
  })
  
  function formatDate(dateString) {
    if (!dateString) return ''
    try {
      return format(new Date(dateString), 'Pp')
    } catch (e) {
      return dateString
    }
  }
  </script>
  
  <style scoped>
  li {
    cursor: pointer;
  }
  .conversation-photo {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    object-fit: cover;
  }
  </style>
  