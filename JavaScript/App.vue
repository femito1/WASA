//<script>
export default {
    data() {
        return {
            count: 0}
        },


    methods:{
        incr(){
            this.count++;
        },
        decr(){
            this.count--;
        }
    }
}
</script>

<template>
    <button @click="incr()"> + </button>
    <button @click="">{{ count }} </button>
    <button @click="decr()"> - </button>
</template>

<style scoped>
button { font-weight: bold; }
</style>