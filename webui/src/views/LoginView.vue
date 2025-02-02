<script>
import axios from 'axios';

export default {
  data() {
    return {
      username: '', // Campo per l'username
      errorMsg: null, // Messaggio di errore
      loading: false, // Stato di caricamento
      userId: null, // ID utente dopo il login
    };
  },
  methods: {
    async handleLogin() {
      this.errorMsg = null;
      this.loading = true;

      // Validazione lato frontend
      if (this.username.length < 3 || this.username.length > 50) {
        this.errorMsg = 'Username must be between 3 and 50 characters.';
        this.loading = false;
        return;
      }

      try {
        // Richiesta API al backend
        const response = await axios.post('/session', { username: this.username });
        
        // Salva l'ID utente restituito
        this.userId = response.data.user_id;
        sessionStorage.setItem('user_id', this.userId); // Salva nel sessionStorage

        // Reindirizza l'utente
        this.$router.push('/dashboard');
      } catch (error) {
        // Gestione degli errori
        this.errorMsg = error.response?.data || 'Login failed. Please try again.';
      } finally {
        this.loading = false;
      }
    },
  },
};
</script>

<template>
  <div class="login-container">
    <!-- Messaggio di errore -->
    <h3 v-if="errorMsg" class="alert alert-danger">{{ errorMsg }}</h3>

    <!-- Indicatore di caricamento -->
    <div v-if="loading">Loading...</div>

    <!-- Form di login -->
    <form @submit.prevent="handleLogin">
      <label for="username">Username:</label>
      <input
        type="text"
        id="username"
        v-model="username"
        required
        minlength="3"
        maxlength="50"
        placeholder="Enter your username"
      />
      <button type="submit">Login</button>
    </form>

    <!-- Messaggio di benvenuto -->
    <div v-if="userId">
      <p>Welcome, {{ username }}!</p>
    </div>
  </div>
</template>

<style>
/* Stili di base per il login */
.login-container {
  max-width: 400px;
  margin: 0 auto;
  padding: 20px;
  border: 1px solid #ddd;
  border-radius: 8px;
  background-color: #f9f9f9;
}

input {
  display: block;
  width: 100%;
  margin: 10px 0;
  padding: 8px;
}

button {
  padding: 10px 15px;
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

button:hover {
  background-color: #0056b3;
}
</style>
