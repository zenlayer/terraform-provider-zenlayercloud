package zenlayer

import (
        "bytes"
        "fmt"
        "hash/crc32"
)

func String(s string) int {
        v := int(crc32.ChecksumIEEE([]byte(s)))
        if v >= 0 {
                return v
        }
        if -v >= 0 {
                return -v
        }
        // v == MinInt
        return 0
}

func dataResourceIdHash(ids []string) string {
        var buf bytes.Buffer

        for _, id := range ids {
                buf.WriteString(id)
        }

        return fmt.Sprintf("%d", String(buf.String()))
}
