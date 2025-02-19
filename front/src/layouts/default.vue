<template>
  <!-- 默认布局 -->
  <main class="h-full w-full overflow-auto bg-[#F5F7FA]">
    <PageHeader class="w-full sticky top-0 z-1" :title="$route.meta?.title" :back="$route.meta?.back" />
    <RouterView />
  </main>
</template>
