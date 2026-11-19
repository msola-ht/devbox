package systeminfo

import (
	"context"
	"encoding/xml"
	"os/exec"
	"strconv"
	"strings"
)

func physicalDiskCapacity(ctx context.Context) *uint64 {
	output, err := exec.CommandContext(ctx, "/usr/sbin/diskutil", "list", "-plist", "physical").Output()
	if err != nil {
		return nil
	}

	var plist diskPlistValue
	if err := xml.Unmarshal(output, &plist); err != nil || len(plist.Children) != 1 {
		return nil
	}
	disks := plist.Children[0].entry("AllDisksAndPartitions")
	if disks == nil || disks.XMLName.Local != "array" {
		return nil
	}

	var total uint64
	for _, disk := range disks.Children {
		size := disk.entry("Size")
		if size == nil || size.XMLName.Local != "integer" {
			return nil
		}
		capacity, err := strconv.ParseUint(strings.TrimSpace(size.Text), 10, 64)
		if err != nil {
			return nil
		}
		total += capacity
	}
	if total == 0 {
		return nil
	}
	return &total
}

type diskPlistValue struct {
	XMLName  xml.Name
	Text     string           `xml:",chardata"`
	Children []diskPlistValue `xml:",any"`
}

func (value *diskPlistValue) entry(key string) *diskPlistValue {
	if value.XMLName.Local != "dict" {
		return nil
	}
	for i := 0; i+1 < len(value.Children); i += 2 {
		name := value.Children[i]
		if name.XMLName.Local == "key" && strings.TrimSpace(name.Text) == key {
			return &value.Children[i+1]
		}
	}
	return nil
}
