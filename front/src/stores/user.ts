import { defineStore } from 'pinia';
import { ref } from 'vue';

export default defineStore('user', () => {
  const user = ref<IUser>();
  const getUser = async () => {
    const username = window.PROJECT_CONFIG.LOGIN_NAME || '';
    user.value = {
      username,
      avatar_url: username ? `${window.BK_DAYU_HOST}/avatars/${username}/avatar.jpg` : '',
    };
  };

  return {
    user,
    getUser,
  };
});
