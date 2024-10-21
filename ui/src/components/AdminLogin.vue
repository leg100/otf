<template>
  <div class="m-auto">
    <form class="flex flex-col gap-2" action="/admin/login" method="POST">
      <div class="field">
        <label for="token">Site Admin Token</label>
        <input class="text-input w-80" type="password" name="token" id="token" required>
      </div>
      <div>
        <button class="btn">Login</button>
      </div>
    </form>
  </div>
</template>
