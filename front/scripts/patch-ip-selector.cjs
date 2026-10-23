/* eslint-disable */
/**
 * 修复 @blueking/ip-selector 静态拓扑面板「取消全选后右侧结果预览不清空」的问题。
 *
 * 根因：handlePageCheck 的取消分支（pageCancle/allCancle）只更新内部 hostCheckedMap，
 * 不触发 change 事件，右侧结果预览（ResultPreview）和外部 value 均无法同步。
 *
 * 修复：取消全选时直接清空已选主机列表，并触发 change 同步预览。
 */
const fs = require('fs');
const path = require('path');

const candidateFiles = ['vue2.6.x.esm.js', 'vue2.6.x.js', 'index.esm.js', 'index.js']
  .map(name => path.join(__dirname, `../node_modules/@blueking/ip-selector/dist/${name}`));

const OLD = `      }).then(function () {
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
      });`;

const NEW = `      }).then(function () {
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
      });`;

let patchedCount = 0;

for (const targetFile of candidateFiles) {
  if (!fs.existsSync(targetFile)) {
    continue;
  }

  const content = fs.readFileSync(targetFile, 'utf8');

  // 幂等：已打过补丁则跳过
  if (content.includes('patch-ip-selector')) {
    console.log(`[patch-ip-selector] 补丁已应用，跳过: ${path.basename(targetFile)}`);
    patchedCount += 1;
    continue;
  }

  if (!content.includes(OLD)) {
    console.warn(
      `[patch-ip-selector] ${path.basename(targetFile)} 未匹配到目标代码，版本可能已更新，请手动检查取消全选逻辑`,
    );
    continue;
  }

  fs.writeFileSync(targetFile, content.replace(OLD, NEW), 'utf8');
  patchedCount += 1;
  console.log(`[patch-ip-selector] 补丁应用成功: ${path.basename(targetFile)}`);
}

if (patchedCount === 0) {
  console.warn('[patch-ip-selector] 未对任何文件应用补丁，请确认 @blueking/ip-selector 已安装');
}
