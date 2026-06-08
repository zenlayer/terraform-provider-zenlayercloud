# ============================================================
# Example 1: 最简配置 — 仅策略名称
# ============================================================
resource "zenlayercloud_zec_ddos_policy" "basic" {
  policy_name = "basic-ddos-policy"
}

# ============================================================
# Example 2: 绑定 EIP（先创建 EIP，再绑定到策略）
# ============================================================
resource "zenlayercloud_zec_eip" "example" {
  region_id            = "SEA-A"
  name                 = "example-eip"
  internet_charge_type = "ByBandwidth"
  bandwidth            = 10
}

resource "zenlayercloud_zec_ddos_policy" "with_eip" {
  policy_name  = "policy-with-eip"
  ipv4_id_list = [zenlayercloud_zec_eip.example.id]
}

# ============================================================
# Example 3: IP 黑白名单 + 超时
# ============================================================
resource "zenlayercloud_zec_ddos_policy" "ip_list" {
  policy_name = "ip-list-policy"

  black_ip_list    = ["1.2.3.4", "5.6.7.0/24"]
  white_ip_list    = ["10.0.0.1", "192.168.1.0/24"]
  ip_black_timeout = 60 # 60 分钟后自动解封
}

# ============================================================
# Example 4: 协议封禁（封禁 ICMP，不能同时封禁 TCP 和 UDP）
# ============================================================
resource "zenlayercloud_zec_ddos_policy" "block_protocol" {
  policy_name    = "block-protocol-policy"
  block_protocol = ["ICMP"]
}

# ============================================================
# Example 5: 区域封禁
# ============================================================
resource "zenlayercloud_zec_ddos_policy" "block_region" {
  policy_name   = "block-region-policy"
  block_regions = ["BTN", "CHN"]
}

# ============================================================
# Example 6: 端口封禁规则
# ============================================================
resource "zenlayercloud_zec_ddos_policy" "port_block" {
  policy_name = "port-block-policy"

  # 封禁所有 UDP 53 端口（DNS 反射攻击防护）
  port {
    protocol       = "UDP"
    src_port_start = 0
    src_port_end   = 65535
    dst_port_start = 53
    dst_port_end   = 53
    action         = "Drop"
  }

  # 封禁 TCP 445 端口（SMB 攻击防护）
  port {
    protocol       = "TCP"
    src_port_start = 0
    src_port_end   = 65535
    dst_port_start = 445
    dst_port_end   = 445
    action         = "Drop"
  }
}

# ============================================================
# Example 7: UDP 反射攻击防护端口
# ============================================================
resource "zenlayercloud_zec_ddos_policy" "reflect_udp" {
  policy_name = "reflect-udp-policy"

  # 在系统默认封禁端口之外，额外追加自定义反射攻击源端口
  reflect_udp_port {
    port = 10001
  }
  reflect_udp_port {
    port = 20001
  }
}

# ============================================================
# Example 8: 指纹过滤规则
# ============================================================
resource "zenlayercloud_zec_ddos_policy" "fingerprint" {
  policy_name = "fingerprint-policy"

  fingerprint_rule {
    protocol       = "UDP"
    src_port_start = 0
    src_port_end   = 65535
    dst_port_start = 0
    dst_port_end   = 65535
    min_pkt_length = 64
    max_pkt_length = 1500
    match_bytes    = "deadbeef" # 16进制小写，不含 0x 前缀
    offset         = 0
    action         = "Drop"
  }
}

# ============================================================
# Example 9: 源限速配置
# ============================================================
resource "zenlayercloud_zec_ddos_policy" "traffic_control" {
  policy_name = "traffic-control-policy"

  traffic_control {
    bps_enabled = true
    bps         = 104857600 # 100 Mbps（单位 bps）

    pps_enabled = true
    pps         = 50000 # 最大值 50000 pps

    syn_bps_enabled = true
    syn_bps         = 10485760 # 10 Mbps

    syn_pps_enabled = true
    syn_pps         = 10000 # 1w pps
  }
}

# ============================================================
# Example 10: 完整配置（含 EIP 绑定、多种防护规则和标签）
# ============================================================
resource "zenlayercloud_zec_eip" "full" {
  region_id            = "SEA-A"
  name                 = "full-example-eip"
  internet_charge_type = "ByBandwidth"
  bandwidth            = 100
}

resource "zenlayercloud_zec_ddos_policy" "full" {
  policy_name       = "full-ddos-policy"
  resource_group_id = "rg-xxxxxxxx"

  # 绑定 EIP
  ipv4_id_list = [zenlayercloud_zec_eip.full.id]

  # IP 黑白名单
  black_ip_list    = ["1.2.3.4"]
  white_ip_list    = ["10.0.0.1"]
  ip_black_timeout = 120

  # 协议封禁
  block_protocol = ["ICMP"]

  # 区域封禁
  block_regions = ["BTN", "CHN"]

  # 端口封禁
  port {
    protocol       = "UDP"
    src_port_start = 0
    src_port_end   = 65535
    dst_port_start = 53
    dst_port_end   = 53
    action         = "Drop"
  }

  # 反射攻击防护
  reflect_udp_port {
    port = 10001
  }
  reflect_udp_port {
    port = 20001
  }

  # 指纹过滤
  fingerprint_rule {
    protocol       = "UDP"
    src_port_start = 0
    src_port_end   = 65535
    dst_port_start = 0
    dst_port_end   = 65535
    min_pkt_length = 100
    max_pkt_length = 1500
    match_bytes    = "cafebabe"
    offset         = 0
    action         = "Drop"
  }

  # 源限速
  traffic_control {
    bps_enabled     = true
    bps             = 104857600
    pps_enabled     = true
    pps             = 10000
    syn_bps_enabled = false
    syn_pps_enabled = false
  }

  tags = {
    env     = "prod"
    project = "security"
  }
}

# ============================================================
# Example 11: 策略更新（ModifyPolicy）
# ============================================================
# 不同字段对应不同 configType，provider 自动按变更字段分发调用：
#   policy_name                                    → 直接更新名称（无 configType）
#   black_ip_list / white_ip_list / ip_black_timeout → IpList
#   block_protocol                                 → BlockProtocol
#   block_regions                                  → BlockRegion
#   port                                           → Port
#   fingerprint_rule                               → Fingerprint
#   reflect_udp_port                               → UdpReflect
#   traffic_control                                → TrafficControl
#   tags                                           → ZRM ModifyResourceTags

resource "zenlayercloud_zec_ddos_policy" "update_example" {
  policy_name = "update-demo-policy"

  # 修改任意字段后执行 terraform apply，provider 自动调用对应 configType 的 ModifyPolicy
  block_protocol = ["ICMP"]

  black_ip_list    = ["1.2.3.4"]
  ip_black_timeout = 60

  tags = {
    env = "test"
  }
}

# ============================================================
# Example 12: EIP 换绑（depends_on 保证先解绑再换绑）
# ============================================================
# 限制：一个 EIP 同一时间只能属于一个策略。
# 换绑时 provider 会先调 DetachFromPolicy，再调 AttachToPolicy。
# Terraform 默认并发更新资源，必须通过 depends_on 保证顺序。

resource "zenlayercloud_zec_ddos_policy" "rebind_policy_a" {
  policy_name = "rebind-policy-a"
  # Step 1：ipv4_id_list 填入 EIP ID，apply 后 EIP 绑定在此策略
  # Step 2：清空 ipv4_id_list，apply 触发 DetachFromPolicy
}

resource "zenlayercloud_zec_ddos_policy" "rebind_policy_b" {
  policy_name = "rebind-policy-b"
  # Step 2：填入 EIP ID，apply 触发 AttachToPolicy
  # ipv4_id_list = ["<eip-id>"]

  # 必须声明 depends_on，确保 policy_a 先解绑完成再换绑
  depends_on = [zenlayercloud_zec_ddos_policy.rebind_policy_a]
}

# ============================================================
# Example 13: 数据源 — 查询所有策略
# ============================================================
data "zenlayercloud_zec_ddos_policies" "all" {
}

output "all_policies" {
  value = data.zenlayercloud_zec_ddos_policies.all.result
}

# ============================================================
# Example 14: 数据源 — 按名称模糊搜索
# ============================================================
data "zenlayercloud_zec_ddos_policies" "by_name" {
  policy_name = "prod"
}

output "prod_policies" {
  value = data.zenlayercloud_zec_ddos_policies.by_name.result
}

# ============================================================
# Example 15: 数据源 — 按 ID 查询指定策略
# ============================================================
data "zenlayercloud_zec_ddos_policies" "by_ids" {
  policy_ids = [
    zenlayercloud_zec_ddos_policy.full.id,
    zenlayercloud_zec_ddos_policy.basic.id,
  ]
}

output "specific_policies" {
  value = data.zenlayercloud_zec_ddos_policies.by_ids.result
}

# ============================================================
# Example 16: 数据源 — 导出结果到文件
# ============================================================
data "zenlayercloud_zec_ddos_policies" "export" {
  result_output_file = "./ddos_policies.json"
}
