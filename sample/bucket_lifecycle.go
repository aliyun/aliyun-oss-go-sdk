package sample

import (
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// BucketLifecycleSample shows how to set, get and delete bucket's lifecycle.
func BucketLifecycleSample() {
	// New client
	client, err := oss.New(endpoint, accessID, accessKey)
	if err != nil {
		HandleError(err)
	}

	// Create the bucket with default parameters
	err = client.CreateBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Case 1: Set the lifecycle. The rule ID is rule1 and the applied objects' prefix is one and the last modified Date is before 2015/11/11
	expiration := oss.LifecycleExpiration{
		CreatedBeforeDate: "2015-11-11T00:00:00.000Z",
	}
	rule1 := oss.LifecycleRule{
		ID:         "rule1",
		Prefix:     "one",
		Status:     "Enabled",
		Expiration: &expiration,
	}
	var rules = []oss.LifecycleRule{rule1}
	err = client.SetBucketLifecycle(bucketName, rules)
	if err != nil {
		HandleError(err)
	}

	// Case 2: Get the bucket's lifecycle
	lc, err := client.GetBucketLifecycle(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("Bucket Lifecycle:%v, %v\n", lc.Rules, *lc.Rules[0].Expiration)

	// Case 3: Set the lifecycle, The rule ID is rule2 and the applied objects' prefix is two. The object start with the prefix will be transited to IA storage Type 3 days latter, and to archive storage type 30 days latter
	transitionIA := oss.LifecycleTransition{
		Days:         3,
		StorageClass: oss.StorageIA,
	}
	transitionArch := oss.LifecycleTransition{
		Days:         30,
		StorageClass: oss.StorageArchive,
	}
	rule2 := oss.LifecycleRule{
		ID:          "rule2",
		Prefix:      "two",
		Status:      "Enabled",
		Transitions: []oss.LifecycleTransition{transitionIA, transitionArch},
	}
	rules = []oss.LifecycleRule{rule2}
	err = client.SetBucketLifecycle(bucketName, rules)
	if err != nil {
		HandleError(err)
	}

	// Case 4: Set the lifecycle, The rule ID is rule3 and the applied objects' prefix is three. The object start with the prefix will be transited to IA storage Type 3 days latter, and to archive storage type 30 days latter, the uncompleted multipart upload will be abort 3 days latter.
	abortMPU := oss.LifecycleAbortMultipartUpload{
		Days: 3,
	}
	rule3 := oss.LifecycleRule{
		ID:                   "rule3",
		Prefix:               "three",
		Status:               "Enabled",
		AbortMultipartUpload: &abortMPU,
	}
	rules = append(lc.Rules, rule3)
	err = client.SetBucketLifecycle(bucketName, rules)
	if err != nil {
		HandleError(err)
	}

	// Case 5: Set the lifecycle. The rule ID is rule4 and the applied objects' has the tagging which prefix is four and the last modified Date is before 2015/11/11
	expiration = oss.LifecycleExpiration{
		CreatedBeforeDate: "2015-11-11T00:00:00.000Z",
	}
	tag1 := oss.Tag{
		Key:   "key1",
		Value: "value1",
	}
	tag2 := oss.Tag{
		Key:   "key2",
		Value: "value2",
	}
	rule4 := oss.LifecycleRule{
		ID:         "rule4",
		Prefix:     "four",
		Status:     "Enabled",
		Tags:       []oss.Tag{tag1, tag2},
		Expiration: &expiration,
	}
	rules = []oss.LifecycleRule{rule4}
	err = client.SetBucketLifecycle(bucketName, rules)
	if err != nil {
		HandleError(err)
	}

	// Case 6: Set the lifecycle. The rule ID is filter one and Include Not exclusion conditions
	expiration = oss.LifecycleExpiration{
		CreatedBeforeDate: "2015-11-11T00:00:00.000Z",
	}
	tag := oss.Tag{
		Key:   "key1",
		Value: "value1",
	}
	greater := int64(500)
	less := int64(645000)
	filter := oss.LifecycleFilter{
		ObjectSizeLessThan:    &greater,
		ObjectSizeGreaterThan: &less,
		Not: []oss.LifecycleFilterNot{
			{
				Prefix: "logs/log2",
				Tag:    &tag,
			},
		},
	}
	filterRule := oss.LifecycleRule{
		ID:         "filter one",
		Prefix:     "logs",
		Status:     "Enabled",
		Expiration: &expiration,
		Transitions: []oss.LifecycleTransition{
			{
				Days:         10,
				StorageClass: oss.StorageIA,
			},
		},
		Filter: &filter,
	}
	rules = []oss.LifecycleRule{filterRule}
	err = client.SetBucketLifecycle(bucketName, rules)
	if err != nil {
		HandleError(err)
	}

	// Case 7: Set the lifecycle. The rules with amtime and return to std when visit
	isTrue := true
	isFalse := false
	rule1 = oss.LifecycleRule{
		ID:     "mtime transition1",
		Prefix: "logs1",
		Status: "Enabled",
		Transitions: []oss.LifecycleTransition{
			{
				Days:         30,
				StorageClass: oss.StorageIA,
			},
		},
	}
	rule2 = oss.LifecycleRule{
		ID:     "mtime transition2",
		Prefix: "logs2",
		Status: "Enabled",
		Transitions: []oss.LifecycleTransition{
			{
				Days:         30,
				StorageClass: oss.StorageIA,
				IsAccessTime: &isFalse,
			},
		},
	}
	rule3 = oss.LifecycleRule{
		ID:     "amtime transition1",
		Prefix: "logs3",
		Status: "Enabled",
		Transitions: []oss.LifecycleTransition{
			{
				Days:                 30,
				StorageClass:         oss.StorageIA,
				IsAccessTime:         &isTrue,
				ReturnToStdWhenVisit: &isFalse,
			},
		},
	}
	rule4 = oss.LifecycleRule{
		ID:     "amtime transition2",
		Prefix: "logs4",
		Status: "Enabled",
		Transitions: []oss.LifecycleTransition{
			{
				Days:                 30,
				StorageClass:         oss.StorageIA,
				IsAccessTime:         &isTrue,
				ReturnToStdWhenVisit: &isTrue,
				AllowSmallFile:       &isFalse,
			},
		},
	}
	rule5 := oss.LifecycleRule{
		ID:     "amtime transition3",
		Prefix: "logs5",
		Status: "Enabled",
		NonVersionTransitions: []oss.LifecycleVersionTransition{
			{
				NoncurrentDays:       10,
				StorageClass:         oss.StorageIA,
				IsAccessTime:         &isTrue,
				ReturnToStdWhenVisit: &isFalse,
				AllowSmallFile:       &isTrue,
			},
		},
	}
	rules = []oss.LifecycleRule{rule1, rule2, rule3, rule4, rule5}
	err = client.SetBucketLifecycle(bucketName, rules)
	if err != nil {
		HandleError(err)
	}

	// case 8: Set bucket's Lifecycle with xml
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<LifecycleConfiguration>
	<Rule>
    <ID>mtime transition1</ID>
    <Prefix>logs1/</Prefix>
    <Status>Enabled</Status>
    <Transition>
      <Days>30</Days>
      <StorageClass>IA</StorageClass>
    </Transition>
  </Rule>
  <Rule>
    <ID>mtime transition2</ID>
    <Prefix>logs2/</Prefix>
    <Status>Enabled</Status>
    <Transition>
      <Days>30</Days>
      <StorageClass>IA</StorageClass>
      <IsAccessTime>false</IsAccessTime>
    </Transition>
  </Rule>
  <Rule>
    <ID>atime transition1</ID>
    <Prefix>logs3/</Prefix>
    <Status>Enabled</Status>
    <Transition>
      <Days>30</Days>
      <StorageClass>IA</StorageClass>
      <IsAccessTime>true</IsAccessTime>
      <ReturnToStdWhenVisit>false</ReturnToStdWhenVisit>
    </Transition>
  </Rule>
  <Rule>
    <ID>atime transition2</ID>
    <Prefix>logs4/</Prefix>
    <Status>Enabled</Status>
    <Transition>
      <Days>30</Days>
      <StorageClass>IA</StorageClass>
      <IsAccessTime>true</IsAccessTime>
      <ReturnToStdWhenVisit>true</ReturnToStdWhenVisit>
      <AllowSmallFile>false</AllowSmallFile>
    </Transition>
  </Rule>
  <Rule>
    <ID>atime transition3</ID>
    <Prefix>logs5/</Prefix>
    <Status>Enabled</Status>
    <NoncurrentVersionTransition>
      <NoncurrentDays>10</NoncurrentDays>
      <StorageClass>IA</StorageClass>
      <IsAccessTime>true</IsAccessTime>
      <ReturnToStdWhenVisit>false</ReturnToStdWhenVisit>
	  <AllowSmallFile>true</AllowSmallFile>
    </NoncurrentVersionTransition>
  </Rule>
  <Rule>
    <ID>r1</ID>
    <Prefix>abc/</Prefix>
    <Filter>
      <ObjectSizeGreaterThan>500</ObjectSizeGreaterThan>
      <ObjectSizeLessThan>64000</ObjectSizeLessThan>
      <Not>
        <Prefix>abc/not1/</Prefix>
        <Tag>
          <Key>notkey1</Key>
          <Value>notvalue1</Value>
        </Tag>
      </Not>
      <Not>
        <Prefix>abc/not2/</Prefix>
        <Tag>
          <Key>notkey2</Key>
          <Value>notvalue2</Value>
        </Tag>
      </Not>
    </Filter>
  </Rule>
</LifecycleConfiguration>
`
	err = client.SetBucketLifecycleXml(bucketName, xmlData)
	if err != nil {
		HandleError(err)
	}

	// case 9: Get bucket's Lifecycle print info
	lcRes, err := client.GetBucketLifecycle(bucketName)
	if err != nil {
		HandleError(err)
	}
	for _, rule := range lcRes.Rules {
		fmt.Println("Lifecycle Rule Id:", rule.ID)
		fmt.Println("Lifecycle Rule Prefix:", rule.Prefix)
		fmt.Println("Lifecycle Rule Status:", rule.Status)
		if rule.Expiration != nil {
			fmt.Println("Lifecycle Rule Expiration Days:", rule.Expiration.Days)
			fmt.Println("Lifecycle Rule Expiration Date:", rule.Expiration.Date)
			fmt.Println("Lifecycle Rule Expiration Created Before Date:", rule.Expiration.CreatedBeforeDate)
			if rule.Expiration.ExpiredObjectDeleteMarker != nil {
				fmt.Println("Lifecycle Rule Expiration Expired Object DeleteMarker:", *rule.Expiration.ExpiredObjectDeleteMarker)
			}
		}

		for _, tag := range rule.Tags {
			fmt.Println("Lifecycle Rule Tag Key:", tag.Key)
			fmt.Println("Lifecycle Rule Tag Value:", tag.Value)
		}

		for _, transition := range rule.Transitions {
			fmt.Println("Lifecycle Rule Transition Days:", transition.Days)
			fmt.Println("Lifecycle Rule Transition Created Before Date:", transition.CreatedBeforeDate)
			fmt.Println("Lifecycle Rule Transition Storage Class:", transition.StorageClass)
			if transition.IsAccessTime != nil {
				fmt.Println("Lifecycle Rule Transition Is Access Time:", *transition.IsAccessTime)
			}
			if transition.ReturnToStdWhenVisit != nil {
				fmt.Println("Lifecycle Rule Transition Return To Std When Visit:", *transition.ReturnToStdWhenVisit)
			}

			if transition.AllowSmallFile != nil {
				fmt.Println("Lifecycle Rule Transition Allow Small File:", *transition.AllowSmallFile)
			}

		}
		if rule.AbortMultipartUpload != nil {
			fmt.Println("Lifecycle Rule Abort Multipart Upload Days:", rule.AbortMultipartUpload.Days)
			fmt.Println("Lifecycle Rule Abort Multipart Upload Created Before Date:", rule.AbortMultipartUpload.CreatedBeforeDate)
		}

		if rule.NonVersionExpiration != nil {
			fmt.Println("Lifecycle Non Version Expiration Non Current Days:", rule.NonVersionExpiration.NoncurrentDays)
		}

		for _, nonVersionTransition := range rule.NonVersionTransitions {
			fmt.Println("Lifecycle Rule Non Version Transitions Non current Days:", nonVersionTransition.NoncurrentDays)
			fmt.Println("Lifecycle Rule Non Version Transition Storage Class:", nonVersionTransition.StorageClass)
			if nonVersionTransition.IsAccessTime != nil {
				fmt.Println("Lifecycle Rule Non Version Transition Is Access Time:", *nonVersionTransition.IsAccessTime)
			}

			if nonVersionTransition.ReturnToStdWhenVisit != nil {
				fmt.Println("Lifecycle Rule Non Version Transition Return To Std When Visit:", *nonVersionTransition.ReturnToStdWhenVisit)
			}

			if nonVersionTransition.AllowSmallFile != nil {
				fmt.Println("Lifecycle Rule Non Version Allow Small File:", *nonVersionTransition.AllowSmallFile)
			}

			if rule.Filter != nil {
				if rule.Filter.ObjectSizeGreaterThan != nil {
					fmt.Println("Lifecycle Rule Filter Object Size Greater Than:", *rule.Filter.ObjectSizeGreaterThan)
				}
				if rule.Filter.ObjectSizeLessThan != nil {
					fmt.Println("Lifecycle Rule Filter Object Size Less Than:", *rule.Filter.ObjectSizeLessThan)
				}
				for _, filterNot := range rule.Filter.Not {
					fmt.Println("Lifecycle Rule Filter Not Prefix:", filterNot.Prefix)
					if filterNot.Tag != nil {
						fmt.Println("Lifecycle Rule Filter Not Tag Key:", filterNot.Tag.Key)
						fmt.Println("Lifecycle Rule Filter Not Tag Value:", filterNot.Tag.Value)
					}
				}
			}
		}
	}

	// Case 10: Delete bucket's Lifecycle
	err = client.DeleteBucketLifecycle(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Delete bucket
	err = client.DeleteBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("BucketLifecycleSample completed")
}
