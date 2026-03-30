<template>
  <el-dialog
    title="S3 密钥浏览器 (S3 Key Browser)"
    :visible.sync="dialogVisible"
    width="500px"
    :modal="false"
    @close="onClose"
  >
    <el-tree
      :props="treeProps"
      :load="loadNode"
      lazy
      node-key="path"
      highlight-current
      @node-click="onNodeClick"
    >
      <span class="custom-tree-node" slot-scope="{ data }">
        <i :class="data.isDir ? 'el-icon-folder' : 'el-icon-document'" style="margin-right: 6px;"></i>
        <span>{{ data.name }}</span>
        <span v-if="!data.isDir && data.size" style="color: #999; margin-left: 8px; font-size: 12px;">
          ({{ formatSize(data.size) }})
        </span>
      </span>
    </el-tree>
    <div v-if="selectedPath" style="margin-top: 12px; padding: 8px; background: #f5f7fa; border-radius: 4px; word-break: break-all;">
      <span style="color: #606266; font-size: 13px;">已选择: {{ selectedPath }}</span>
    </div>
    <span slot="footer">
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" :disabled="!selectedPath" @click="onConfirm">确定</el-button>
    </span>
  </el-dialog>
</template>

<script>
import { listS3Objects } from '@/api/s3'

export default {
  props: {
    visible: {
      type: Boolean,
      default: false
    }
  },
  data() {
    return {
      selectedPath: '',
      treeProps: {
        label: 'name',
        isLeaf: (data) => !data.isDir,
        children: 'children'
      }
    }
  },
  computed: {
    dialogVisible: {
      get() { return this.visible },
      set(val) { this.$emit('update:visible', val) }
    }
  },
  methods: {
    async loadNode(node, resolve) {
      const prefix = node.level === 0 ? '' : node.data.path
      try {
        const res = await listS3Objects(prefix)
        if (res.Msg === 'success' && res.Data) {
          resolve(res.Data)
        } else {
          resolve([])
          if (res.Msg !== 'success') {
            this.$message.error(res.Msg)
          }
        }
      } catch (e) {
        resolve([])
        this.$message.error('加载 S3 目录失败')
      }
    },
    onNodeClick(data) {
      if (!data.isDir) {
        this.selectedPath = data.path
      }
    },
    onConfirm() {
      if (this.selectedPath) {
        this.$emit('select', this.selectedPath)
        this.dialogVisible = false
      }
    },
    onClose() {
      this.selectedPath = ''
    },
    formatSize(bytes) {
      if (bytes === 0) return '0 B'
      const k = 1024
      const sizes = ['B', 'KB', 'MB', 'GB']
      const i = Math.floor(Math.log(bytes) / Math.log(k))
      return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
    }
  }
}
</script>

<style scoped>
.custom-tree-node {
  display: flex;
  align-items: center;
  font-size: 14px;
}
</style>
