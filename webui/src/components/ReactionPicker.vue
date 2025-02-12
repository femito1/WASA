<!-- File: webui/src/components/ReactionPicker.vue -->
<template>
    <div class="modal-overlay">
      <div class="modal-content">
        <h5>Select a Reaction</h5>
        <div class="emoji-list">
          <span v-for="emoji in emojis" :key="emoji" class="emoji" @click="selectReaction(emoji)">
            {{ emoji }}
          </span>
        </div>
        <button class="btn btn-secondary mt-2" @click="emitClose">Cancel</button>
      </div>
    </div>
  </template>
  
  <script setup>
  import { ref } from 'vue'
  const props = defineProps({
    message: { type: Object, required: true }
  })
  // Declare emitted events.
  const emit = defineEmits(['react','close'])
  // Define a list of emoji reactions.
  const emojis = ref(['👍', '❤️', '😂', '😮', '😢', '👏'])
  
  function selectReaction(emoji) {
    // Emit the reaction event with the selected emoji.
    emit('react', emoji)
  }
  function emitClose() {
    emit('close')
  }
  </script>
  
  <style scoped>
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0,0,0,0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }
  .modal-content {
    background: white;
    padding: 16px;
    border-radius: 8px;
    text-align: center;
  }
  .emoji-list {
    display: flex;
    justify-content: center;
    gap: 12px;
    margin-top: 8px;
  }
  .emoji {
    font-size: 1.5rem;
    cursor: pointer;
  }
  </style>
  