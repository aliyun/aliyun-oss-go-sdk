package sample

import (
	"encoding/xml"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketReplicationSample  how to set, get or delete the bucket replication.
func BucketReplicationSample() {
	// New client
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		HandleError(err)
	}

	// Create a bucket with default parameters
	err = client.CreateBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Case 1:Put Bucket Replication
	// Case 1-1:Put Bucket Replication in xml format
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<ReplicationConfiguration>
 <Rule>
    <PrefixSet>
       <Prefix>source1</Prefix>
       <Prefix>video</Prefix>
    </PrefixSet>
    <Action>PUT</Action>
    <Destination>
       <Bucket>destBucketName</Bucket>
       <Location>oss-cn-hangzhou</Location>
       <TransferType>oss_acc</TransferType>
    </Destination>
    <HistoricalObjectReplication>enabled</HistoricalObjectReplication>
     <SyncRole>aliyunramrole</SyncRole>
     <SourceSelectionCriteria>
        <SseKmsEncryptedObjects>
          <Status>Enabled</Status>
        </SseKmsEncryptedObjects>
     </SourceSelectionCriteria>
     <EncryptionConfiguration>
          <ReplicaKmsKeyID>c4d49f85-ee30-426b-a5ed-95e9139d****</ReplicaKmsKeyID>
     </EncryptionConfiguration>
 </Rule>
</ReplicationConfiguration>`

	err = client.PutBucketReplication(bucketName, xmlData)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Put Bucket Replication in xml format Success!")
	// Case 1-2:Put Bucket Replication in Struct
	destBucketName := "yp-re"
	prefix1 := "prefix_1"
	prefix2 := "prefix_2"
	keyId := "c4d49f85-ee30-426b-a5ed-95e9139d******"
	source := "Enabled"
	prefixSet := oss.ReplicationRulePrefix{Prefix: []*string{&prefix1, &prefix2}}
	reqReplication := oss.PutBucketReplication{
		Rule: []oss.ReplicationRule{
			{
				PrefixSet: &prefixSet,
				Action:    "ALL",
				Destination: &oss.ReplicationRuleDestination{
					Bucket:       destBucketName,
					Location:     "oss-cn-hangzhou",
					TransferType: "oss_acc",
				},
				HistoricalObjectReplication: "disabled",
				SyncRole:                    "aliyunramrole",
				EncryptionConfiguration:     &keyId,
				SourceSelectionCriteria:     &source,
			},
		},
	}
	xmlBody, err := xml.Marshal(reqReplication)
	if err != nil {
		HandleError(err)
	}
	err = client.PutBucketReplication(bucketName, string(xmlBody))

	if err != nil {
		HandleError(err)
	}
	fmt.Println("Put Bucket Replication Success!")

	// Case 2:Get Bucket Replication
	stringData, err := client.GetBucketReplication(bucketName)
	if err != nil {
		HandleError(err)
	}

	var repResult oss.GetBucketReplicationResult
	err = xml.Unmarshal([]byte(stringData),&repResult)
	if err != nil {
		HandleError(err)
	}
	for _, rule := range repResult.Rule {
		fmt.Printf("Rule Id:%s\n", rule.ID)
		if rule.RTC != nil {
			fmt.Printf("Rule RTC:%s\n", *rule.RTC)
		}
		if rule.PrefixSet != nil {
			for _, prefix := range rule.PrefixSet.Prefix {
				fmt.Printf("Rule Prefix:%s\n", *prefix)
			}
		}
		fmt.Printf("Rule Action:%s\n", rule.Action)
		fmt.Printf("Rule Destination Bucket:%s\n", rule.Destination.Bucket)
		fmt.Printf("Rule Destination Location:%s\n", rule.Destination.Location)
		fmt.Printf("Rule Destination TransferType:%s\n", rule.Destination.TransferType)
		fmt.Printf("Rule Status:%s\n", rule.Status)
		fmt.Printf("Rule Historical Object Replication:%s\n", rule.HistoricalObjectReplication)
		if rule.SyncRole != "" {
			fmt.Printf("Rule SyncRole:%s\n", rule.SyncRole)
		}
	}

	// Case 3:Put Bucket RTC
	enabled := "enabled"
	ruleId := "564df6de-7372-46dc-b4eb-10f******"
	rtc := oss.PutBucketRTC{
		RTC: &enabled,
		ID:  ruleId,
	}
	err = client.PutBucketRTC(bucketName, rtc)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("Put Bucket RTC Success!")
	// Case 4:Get Bucket Replication Location
	stringData, err = client.GetBucketReplicationLocation(bucketName)
	if err != nil {
		HandleError(err)
	}

	var repLocation oss.GetBucketReplicationLocationResult
	err = xml.Unmarshal([]byte(stringData),&repLocation)
	if err != nil {
		HandleError(err)
	}

	for _, location := range repLocation.Location {
		fmt.Printf("Bucket Replication Location: %s\n", location)
	}

	for _, transferType := range repLocation.LocationTransferType {
		fmt.Printf("Bucket Replication Location Transfer Type Location: %s\n", transferType.Location)
		fmt.Printf("Bucket Replication Location Transfer Type Type: %s\n", transferType.TransferTypes)
	}

	for _, rtcLocation := range repLocation.RTCLocation {
		fmt.Printf("Bucket Replication Location RTC Location: %s\n", rtcLocation)
	}
	fmt.Println("Get Bucket Replication Location Success!")
	// Case 5:Get Bucket Replication Progress
	stringData, err = client.GetBucketReplicationProgress(bucketName, ruleId)
	if err != nil {
		HandleError(err)
	}
	var repProgress oss.GetBucketReplicationProgressResult
	err = xml.Unmarshal([]byte(stringData),&repProgress)
	if err != nil {
		HandleError(err)
	}
	for _, repProgressRule := range repProgress.Rule {
		fmt.Printf("Rule Id:%s\n", repProgressRule.ID)
		if repProgressRule.PrefixSet != nil {
			for _, prefix := range repProgressRule.PrefixSet.Prefix {
				fmt.Printf("Rule Prefix:%s\n", *prefix)
			}
		}
		fmt.Printf("Replication Progress Rule Action:%s\n", repProgressRule.Action)
		fmt.Printf("Replication Progress Rule Destination Bucket:%s\n", repProgressRule.Destination.Bucket)
		fmt.Printf("Replication Progress Rule Destination Location:%s\n", repProgressRule.Destination.Location)
		fmt.Printf("Replication Progress Rule Destination TransferType:%s\n", repProgressRule.Destination.TransferType)
		fmt.Printf("Replication Progress Rule Status:%s\n", repProgressRule.Status)
		fmt.Printf("Replication Progress Rule Historical Object Replication:%s\n", repProgressRule.HistoricalObjectReplication)
		if (*repProgressRule.Progress).HistoricalObject != "" {
			fmt.Printf("Replication Progress Rule Progress Historical Object:%s\n", (*repProgressRule.Progress).HistoricalObject)
		}
		fmt.Printf("Replication Progress Rule Progress NewObject:%s\n", (*repProgressRule.Progress).NewObject)
	}
	fmt.Println("Get Bucket Replication Progress Success!")
	// Case 6:Delete Bucket Replication
	err = client.DeleteBucketReplication(bucketName, ruleId)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Delete Bucket Replication Success!")

	fmt.Println("BucketReplicationSample completed")

}
