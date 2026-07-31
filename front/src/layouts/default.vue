<template>
  <!-- 默认布局 -->
  <div class="h-full w-full flex flex-col bg-[#F5F7FA]">
    <PageHeader
      class="w-full flex-shrink-0 z-1"
      :title="$route.meta?.title"
      :sub-title="$route.meta?.subTitle"
      :back="$route.meta?.back" />
    <div class="flex-1 overflow-y-auto min-w-0">
      <RouterView />
    </div>
  </div>
</template>
