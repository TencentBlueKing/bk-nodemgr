/* eslint-disable */
/**
 * @blueking/ip-selector 补丁集合（postinstall 执行，按补丁逐个幂等）。
 *
 * 1. 静态拓扑面板「取消全选后右侧结果预览不清空」
 *    根因：handlePageCheck 的取消分支（pageCancle/allCancle）只更新内部 hostCheckedMap，
 *    不触发 change 事件，右侧结果预览（ResultPreview）和外部 value 均无法同步。
 *    修复：取消全选时直接清空已选主机列表，并触发 change 同步预览。
 *
 * 2. 主机表格默认每页条数随窗口宽度变化（宽屏下为 13 条）
 *    根因：getPaginationDefault 按 floor(height / 41) - 2 推导每页条数，
 *    dialog 模式的 height 按窗口宽度取 570/670/800 三档，宽屏下推导出 13。
 *    修复：默认每页条数固定为 10，与可选项 [10, 20, 50, 100] 保持一致。
 */
const fs = require('fs');
const path = require('path');

const candidateFiles = ['vue2.6.x.esm.js', 'vue2.6.x.js', 'index.esm.js', 'index.js']
  .map(name => path.join(__dirname, `../node_modules/@blueking/ip-selector/dist/${name}`));

const PATCHES = [
  {
    name: 'cancel-select-all',
    marker: 'patch-ip-selector: cancel select-all clears the selected host list',
    old: `      }).then(function () {
        hostCheckedMap.value = checkedMap;
        if (checkValue === 'page' || checkValue === 'all') {
          submitConfirm().then(function () {
            triggerChange();
          }, function () {
            hostCheckedMap.value = lastestCheckedMapCache;
          }).finally(function () {
            syncTablePageCheckValue();
          });
        }
      });`,
    new: `      }).then(function () {
        hostCheckedMap.value = checkedMap;
        if (checkValue === 'page' || checkValue === 'all') {
          submitConfirm().then(function () {
            triggerChange();
          }, function () {
            hostCheckedMap.value = lastestCheckedMapCache;
          }).finally(function () {
            syncTablePageCheckValue();
          });
        } else {
          // patch-ip-selector: cancel select-all clears the selected host list
          hostCheckedMap.value = {};
          triggerChange();
          syncTablePageCheckValue();
        }
      });`,
  },
  {
    name: 'fixed-page-limit',
    marker: 'patch-ip-selector: fixed default page limit',
    old: 'var defaultLimit = Math.floor(height / 41) - 2;',
    new: 'var defaultLimit = 10; // patch-ip-selector: fixed default page limit',
  },
];

let patchedCount = 0;

for (const targetFile of candidateFiles) {
  if (!fs.existsSync(targetFile)) {
    continue;
  }

  const fileName = path.basename(targetFile);
  let content = fs.readFileSync(targetFile, 'utf8');
  let changed = false;

  for (const patch of PATCHES) {
    // 幂等：已打过该补丁则跳过
    if (content.includes(patch.marker)) {
      console.log(`[patch-ip-selector] ${patch.name} 已应用，跳过: ${fileName}`);
      continue;
    }

    if (!content.includes(patch.old)) {
      console.warn(`[patch-ip-selector] ${patch.name} 未匹配到目标代码，版本可能已更新，请手动检查: ${fileName}`);
      continue;
    }

    content = content.replace(patch.old, patch.new);
    changed = true;
    console.log(`[patch-ip-selector] ${patch.name} 应用成功: ${fileName}`);
  }

  if (changed) {
    fs.writeFileSync(targetFile, content, 'utf8');
  }
  patchedCount += 1;
}

if (patchedCount === 0) {
  console.warn('[patch-ip-selector] 未对任何文件应用补丁，请确认 @blueking/ip-selector 已安装');
}
