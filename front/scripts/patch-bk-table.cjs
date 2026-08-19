/* eslint-disable */
/**
 * 修复 @blueking/table filter 面板在「fixed 列 + 横向滚动」时定位飘移的问题。
 *
 * 根因：@blueking/table 将 @blueking/vxe-table 打包进 index.es.min.js，
 * 其 filter 面板定位用 offsetLeft 累加 + scrollLeft 补偿，在有 fixed 列时坐标系错乱。
 *
 * 修复：改用 getBoundingClientRect 视觉坐标，面板 offsetParent 与筛选按钮直接相减，
 * 自动消除滚动/fixed 影响。
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

const content = fs.readFileSync(targetFile, 'utf8');

// 幂等：已打过补丁则跳过
if (content.includes('panelParentRect')) {
  console.log('[patch-bk-table] 补丁已应用，跳过');
  process.exit(0);
}

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

if (!content.includes(OLD)) {
  console.warn(
    '[patch-bk-table] 未匹配到目标代码，@blueking/table 版本可能已更新，请手动检查定位逻辑',
  );
  process.exit(0);
}

fs.writeFileSync(targetFile, content.replace(OLD, NEW), 'utf8');
console.log('[patch-bk-table] 补丁应用成功');
