# 数据库手动重置说明

由于PowerShell环境问题，请手动执行以下SQL语句来重置数据库：

## 方法1：直接执行SQL

连接到MySQL数据库 `t3bot`，执行以下SQL：

```sql
-- 1. 删除所有现有规则
DELETE FROM dragon_rules;

-- 2. 结束所有活跃长龙
UPDATE dragon_alerts SET status = 'ended' WHERE status = 'active';

-- 3. 为所有群组重建默认规则
INSERT INTO dragon_rules (chat_id, pattern_type, attribute_type, threshold, enabled)
SELECT 
    cc.chat_id,
    r.pattern_type,
    r.attribute_type,
    r.threshold,
    TRUE as enabled
FROM chat_configs cc
CROSS JOIN (
    SELECT 'a' as pattern_type, 'size' as attribute_type, 5 as threshold
    UNION ALL SELECT 'a', 'parity', 5
    UNION ALL SELECT 'a', 'sum', 5
    UNION ALL SELECT 'ab', 'size', 2
    UNION ALL SELECT 'ab', 'parity', 2
    UNION ALL SELECT 'ab', 'sum', 2
    UNION ALL SELECT 'abb', 'size', 2
    UNION ALL SELECT 'abb', 'parity', 2
    UNION ALL SELECT 'abb', 'sum', 2
    UNION ALL SELECT 'ab_ac', 'size_parity', 2
    UNION ALL SELECT 'ab_cd', 'size_parity', 2
    UNION ALL SELECT 'abab', 'size_parity', 2
) r
WHERE cc.chat_id < 0;

-- 4. 查看重置结果
SELECT COUNT(*) as total_rules FROM dragon_rules;
SELECT COUNT(*) as total_groups FROM chat_configs WHERE chat_id < 0;
```

## 方法2：使用reset_database.go脚本

如果可以运行Go程序，执行：

```bash
go run reset_database.go
```

## 默认配置说明

重置后，每个群组将有以下默认规则：

| 模式类型 | 属性 | 阈值(组数) | 实际期数 |
|---------|------|-----------|---------|
| a | size | 5 | 5期 |
| a | parity | 5 | 5期 |
| a | sum | 5 | 5期 |
| ab | size | 2 | 4期 |
| ab | parity | 2 | 4期 |
| ab | sum | 2 | 4期 |
| abb | size | 2 | 6期 |
| abb | parity | 2 | 6期 |
| abb | sum | 2 | 6期 |
| ab_ac | size_parity | 2 | 4期 |
| ab_cd | size_parity | 2 | 4期 |
| abab | size_parity | 2 | 4期 |

注意：threshold存储的是"组数"，代码会自动转换为期数进行判断。

