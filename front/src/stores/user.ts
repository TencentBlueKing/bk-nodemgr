import { defineStore } from 'pinia';
import { ref } from 'vue';

export default defineStore('user', () => {
  const user = ref<IUser>();
  const getCookie = (name: string): string | undefined => {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
  
    if (parts.length === 2) {
      const lastPart = parts.pop();
      if (lastPart) {
        return lastPart.split(';').shift();
      }
    }
    return undefined;
  };
  const getUser = async () => {
    const username = getCookie('bk_uid') || '';
    user.value = {
      username,
      avatar_url: `${window.BK_DAYU_HOST}/avatars/${username}/avatar.jpg`,
    };
  };

  return {
    user,
    getUser,
  };
});
