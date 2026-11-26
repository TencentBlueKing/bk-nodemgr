/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 节点管理 (BlueKing - Node Management) available.
 * Copyright (C) Tencent. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.

 * We undertake not to change the open source license (MIT license) applicable

 * to the current version of the project delivered to anyone in the future.
 */

package redisratelimit

// acquireScript uses the same integer credits as the memory implementation.
// Returns zero on acquisition, positive microseconds until the next token on
// denial, -1 for conflicting rules, and -2 for invalid state or clock regression.
const acquireScript = `
local key = KEYS[1]
local refill = tonumber(ARGV[1])
local cost = tonumber(ARGV[2])
local capacity = cost * tonumber(ARGV[3])
local rule = 'token_bucket:' .. table.concat(ARGV, ':')
local clock = redis.call('TIME')
local now = tonumber(clock[1]) * 1000000 + tonumber(clock[2])
local state = redis.call('HMGET', key, 'rule', 'credits', 'last')
local credits = capacity
if state[1] then
    if state[1] ~= rule then
        return -1
    end
    credits = tonumber(state[2])
    local last = tonumber(state[3])
    if not credits or not last or credits < 0 or credits > capacity or now < last then
        return -2
    end
    local elapsed = now - last
    if elapsed >= math.ceil((capacity - credits) / refill) then
        credits = capacity
    else
        credits = credits + elapsed * refill
    end
elseif redis.call('EXISTS', key) ~= 0 then
    return -2
end
local delay = 0
if credits >= cost then
    credits = credits - cost
else
    delay = math.ceil((cost - credits) / refill)
end
-- Format integers explicitly: Lua tostring may round a microsecond Unix timestamp.
redis.call('HSET', key, 'rule', rule, 'credits', string.format('%.0f', credits), 'last', string.format('%.0f', now))
-- Keep the state until it can naturally become full. Add one millisecond for
-- Redis expiration clock granularity.
local ttl = math.ceil(math.ceil((capacity - credits) / refill) / 1000) + 1
redis.call('PEXPIRE', key, ttl)
return delay
`
