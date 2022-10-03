<script>
import axios from "axios";

export default {
  data() {
    return {
      stuff: "",
    };
  },

  methods: {
    callBackend() {
      axios.get("http://localhost:8081/api/db")
        .then(it => this.stuff = it.data);
    },
  },
};
</script>

<template>
  <div class="Backend">
    <h1>This is an about page</h1>
    <button @click="callBackend">click me</button>
    <p>{{ stuff }}</p>
  </div>
</template>
