<!-- File: webui/src/components/CommentItem.vue -->
<template>
    <div class="comment-item d-flex justify-content-between align-items-start">
      <div>
        <strong>{{ comment.senderName || 'Unknown' }}</strong>
        <small class="text-muted ms-2">{{ formatDate(comment.timestamp) }}</small>
        <div>{{ comment.commentText }}</div>
      </div>
      <button class="btn btn-sm btn-link text-danger" @click="$emit('delete', comment.commentId)">Delete</button>
    </div>
  </template>
  
  <script setup>
  import { format } from 'date-fns'
  const props = defineProps({
    comment: {
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
  .comment-item {
    border-bottom: 1px solid #e0e0e0;
    padding: 4px 0;
  }
  </style>
  