/* eslint-disable */
/**
 * 修复 @blueking/table filter 面板在「fixed 列 + 横向滚动」时定位飘移的问题。
 *
 * 根因：@blueking/table 将 @blueking/vxe-table 打包进 index.es.min.js，
 * 其 filter 面板定位用 offsetLeft 累加 + scrollLeft 补偿，在有 fixed 列时坐标系错乱。
 *
 * 修复：改用 getBoundingClientRect 视觉坐标，面板 offsetParent 与筛选按钮直接相减，
 * 自动消除滚动/fixed 影响。
 *
 * 补丁 2：所有列筛选面板顶部增加搜索框（组件本身不支持）。
 * 关键字按选项文本不区分大小写过滤；有关键字时「全选」仅作用于可见选项；
 * 面板关闭或切换列时清空关键字。样式见 src/App.vue 的 .vxe-table--filter-search。
 */
const fs = require('fs');
const path = require('path');

const targetFile = path.join(
  __dirname,
  '../node_modules/@blueking/table/vue3/index.es.min.js',
);

if (!fs.existsSync(targetFile)) {
  console.warn('[patch-bk-table] 未找到 @blueking/table，跳过补丁');
  process.exit(0);
}

let content = fs.readFileSync(targetFile, 'utf8');

const OLD = `            let left2, right2;
            const style = {
              top: \`\${targetElem.offsetTop + targetElem.offsetParent.offsetTop + targetElem.offsetHeight}px\`
            };
            let maxHeight = null;
            const bodyHeight = bodyElem.clientHeight - (headerElem ? headerElem.clientHeight / 2 : 0);
            if (filterHeight >= bodyHeight) {
              maxHeight = Math.max(40, bodyHeight - (filterFootElem ? filterFootElem.offsetHeight : 0) - (filterHeadElem ? filterHeadElem.offsetHeight : 0));
            }
            if (column.fixed === "left") {
              left2 = targetElem.offsetLeft + targetElem.offsetParent.offsetLeft - centerWidth;
            } else if (column.fixed === "right") {
              right2 = targetElem.offsetParent.offsetWidth - targetElem.offsetLeft + (targetElem.offsetParent.offsetParent.offsetWidth - targetElem.offsetParent.offsetLeft) - column.renderWidth - centerWidth;
            } else {
              left2 = targetElem.offsetLeft + targetElem.offsetParent.offsetLeft - centerWidth - bodyElem.scrollLeft;
            }`;

const NEW = `            let left2, right2;
            const panelParentElem = filterWrapperElem.offsetParent;
            const panelParentRect = panelParentElem ? panelParentElem.getBoundingClientRect() : null;
            const targetRect = targetElem.getBoundingClientRect();
            const style = panelParentRect ? {
              top: \`\${targetRect.bottom - panelParentRect.top}px\`
            } : {
              top: \`\${targetElem.offsetTop + targetElem.offsetParent.offsetTop + targetElem.offsetHeight}px\`
            };
            let maxHeight = null;
            const bodyHeight = bodyElem.clientHeight - (headerElem ? headerElem.clientHeight / 2 : 0);
            if (filterHeight >= bodyHeight) {
              maxHeight = Math.max(40, bodyHeight - (filterFootElem ? filterFootElem.offsetHeight : 0) - (filterHeadElem ? filterHeadElem.offsetHeight : 0));
            }
            if (panelParentRect) {
              if (column.fixed === "left") {
                left2 = targetRect.left - panelParentRect.left + targetRect.width / 2 - centerWidth;
              } else if (column.fixed === "right") {
                right2 = panelParentRect.right - targetRect.right + targetRect.width / 2 - centerWidth;
              } else {
                left2 = targetRect.left - panelParentRect.left + targetRect.width / 2 - centerWidth;
              }
            } else {
              if (column.fixed === "left") {
                left2 = targetElem.offsetLeft + targetElem.offsetParent.offsetLeft - centerWidth;
              } else if (column.fixed === "right") {
                right2 = targetElem.offsetParent.offsetWidth - targetElem.offsetLeft + (targetElem.offsetParent.offsetParent.offsetWidth - targetElem.offsetParent.offsetLeft) - column.renderWidth - centerWidth;
              } else {
                left2 = targetElem.offsetLeft + targetElem.offsetParent.offsetLeft - centerWidth - bodyElem.scrollLeft;
              }
            }`;

const applyPatch = (name, marker, pairs) => {
  if (content.includes(marker)) {
    console.log('[patch-bk-table] ' + name + ' 补丁已应用，跳过');
    return;
  }
  for (const [from] of pairs) {
    if (!content.includes(from)) {
      console.warn('[patch-bk-table] ' + name + ' 未匹配到目标代码，@blueking/table 版本可能已更新，请手动检查');
      return;
    }
  }
  for (const [from, to] of pairs) {
    content = content.replace(from, () => to);
  }
  console.log('[patch-bk-table] ' + name + ' 补丁应用成功');
};

applyPatch('filter 定位', 'panelParentRect', [[OLD, NEW]]);

// ---- 补丁 2：筛选面板搜索框 ----
const SEARCH_SETUP_OLD = 'const { computeFilterOpts } = getComputeMaps();';
const SEARCH_SETUP_NEW = SEARCH_SETUP_OLD + `
    const filterSearchKeyword = ref("");
    watch(() => [props.filterStore && props.filterStore.visible, props.filterStore && props.filterStore.column], () => {
      filterSearchKeyword.value = "";
    });
    const getVisibleFilterOptions = () => {
      const { filterStore } = props;
      const keyword = filterSearchKeyword.value.trim().toLowerCase();
      if (!keyword) return filterStore.options;
      return filterStore.options.filter((item) => String(item.label == null ? "" : item.label).toLowerCase().includes(keyword));
    };`;

const CHECK_ALL_OLD = `    const filterCheckAllEvent = (evnt, value) => {
      const { filterStore } = props;
      filterStore.options.forEach((option) => {`;
const CHECK_ALL_NEW = `    const filterCheckAllEvent = (evnt, value) => {
      const { filterStore } = props;
      if (filterSearchKeyword.value.trim()) {
        getVisibleFilterOptions().forEach((option) => {
          option._checked = value;
        });
        $xeTable.checkFilterOptions();
        return;
      }
      filterStore.options.forEach((option) => {`;

const ALL_STATE_OLD = `      const isAllChecked = multiple ? filterStore.isAllSelected : !filterStore.options.some((item) => item._checked);
      const isAllIndeterminate = multiple && filterStore.isIndeterminate;`;
const ALL_STATE_NEW = `      const hasSearchKeyword = !!filterSearchKeyword.value.trim();
      const visibleFilterOptions = getVisibleFilterOptions();
      const isAllChecked = multiple ? (hasSearchKeyword ? visibleFilterOptions.length > 0 && visibleFilterOptions.every((item) => item._checked) : filterStore.isAllSelected) : !filterStore.options.some((item) => item._checked);
      const isAllIndeterminate = multiple && (hasSearchKeyword ? !isAllChecked && visibleFilterOptions.some((item) => item._checked) : filterStore.isIndeterminate);`;

const HEADER_OLD = `class: "vxe-table--filter-header"
        }, [`;
const HEADER_NEW = `class: "vxe-table--filter-header"
        }, [
          h("li", {
            class: "vxe-table--filter-search"
          }, [
            h("input", {
              class: "vxe-table--filter-search-input",
              type: "text",
              value: filterSearchKeyword.value,
              placeholder: /^zh/i.test(document.documentElement.lang || "") ? "搜索" : "Search",
              onInput: (evnt) => {
                filterSearchKeyword.value = evnt.target.value;
              }
            })
          ]),`;

const ALL_CLICK_OLD = 'changeAllOption(evnt, !filterStore.isAllSelected);';
const ALL_CLICK_NEW = 'changeAllOption(evnt, !isAllChecked);';

const BODY_OLD = `}, filterStore.options.map((item) => {
          const isChecked = item._checked;`;
const BODY_NEW = `}, visibleFilterOptions.map((item) => {
          const isChecked = item._checked;`;

applyPatch('filter 搜索框', 'filterSearchKeyword', [
  [SEARCH_SETUP_OLD, SEARCH_SETUP_NEW],
  [CHECK_ALL_OLD, CHECK_ALL_NEW],
  [ALL_STATE_OLD, ALL_STATE_NEW],
  [HEADER_OLD, HEADER_NEW],
  [ALL_CLICK_OLD, ALL_CLICK_NEW],
  [BODY_OLD, BODY_NEW],
]);

fs.writeFileSync(targetFile, content, 'utf8');
