<template>
  <div class="reply-item">
    <div v-if="reply.originalContent" class="reply-preview">
      <blockquote>{{ reply.originalContent }}</blockquote>
    </div>
    <div class="d-flex justify-content-between align-items-start">
      <div>
        <strong>{{ reply.senderName }}</strong>
        <small class="text-muted ms-2">{{ formatDate(reply.timestamp) }}</small>
        <div>{{ reply.replyText }}</div>
      </div>
      <button class="btn btn-sm btn-link text-danger" @click="$emit('delete', reply.id)">Delete</button>
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  reply: { type: Object, required: true }
});

const formatDate = (timestamp) => {
  return new Date(timestamp).toLocaleString();
};
</script>

<style scoped>
.reply-item {
  border-bottom: 1px solid #e0e0e0;
  padding: 4px 0;
}

.reply-preview {
  margin-bottom: 4px;
  padding: 4px;
  background-color: #f1f1f1;
  border-left: 4px solid #ddd;
  font-style: italic;
}
</style> 