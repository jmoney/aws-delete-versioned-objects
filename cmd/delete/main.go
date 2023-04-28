package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var svc = s3.New(session.Must(session.NewSession()))

func DeleteDeleteMarkerPage(versions *s3.ListObjectVersionsOutput, lastPage bool) bool {
	bucket := versions.Name
	objectsToDelete := make([]*s3.ObjectIdentifier, 0)
	now := time.Now().UTC()
	currentBatch := now
	for _, deleteMarker := range versions.DeleteMarkers {
		objectsToDelete = append(objectsToDelete, &s3.ObjectIdentifier{
			Key:       deleteMarker.Key,
			VersionId: deleteMarker.VersionId,
		})

		if currentBatch.After(*deleteMarker.LastModified) {
			currentBatch = *deleteMarker.LastModified
		}
	}

	input := s3.DeleteObjectsInput{
		Bucket: bucket,
		Delete: &s3.Delete{
			Objects: objectsToDelete,
		},
	}

	if len(objectsToDelete) != 0 {
		_, err := svc.DeleteObjects(&input)

		if err != nil {
			panic(err)
		}

		fmt.Printf("Deleted %d delete markers at %s. Earliest in Batch is %s. Continue: %t\n", len(objectsToDelete), now.Format(time.RFC3339), currentBatch.Format(time.RFC3339), !lastPage && (err == nil))
		return !lastPage && (err == nil)
	}
	fmt.Println("Not delete markers to delete.")
	return false
}

func DeleteVersionPage(versions *s3.ListObjectVersionsOutput, lastPage bool) bool {
	bucket := versions.Name
	objectsToDelete := make([]*s3.ObjectIdentifier, 0)
	now := time.Now().UTC()
	currentBatch := now
	for _, version := range versions.Versions {
		objectsToDelete = append(objectsToDelete, &s3.ObjectIdentifier{
			Key:       version.Key,
			VersionId: version.VersionId,
		})

		if currentBatch.After(*version.LastModified) {
			currentBatch = *version.LastModified
		}
	}

	input := s3.DeleteObjectsInput{
		Bucket: bucket,
		Delete: &s3.Delete{
			Objects: objectsToDelete,
		},
	}

	if len(objectsToDelete) != 0 {
		_, err := svc.DeleteObjects(&input)

		if err != nil {
			panic(err)
		}

		fmt.Printf("Deleted %d versions at %s. Earliest in Batch is %s. Continue: %t\n", len(objectsToDelete), now.Format(time.RFC3339), currentBatch.Format(time.RFC3339), !lastPage && (err == nil))
		return !lastPage && (err == nil)
	}
	fmt.Println("No versions to delete.")
	return false
}

func DeletePage(versions *s3.ListObjectVersionsOutput, lastPage bool) bool {
	ret := DeleteVersionPage(versions, lastPage)
	ret = DeleteDeleteMarkerPage(versions, lastPage) || ret
	return ret
}

func main() {
	bucket := flag.String("bucket", "", "The bucket to delete from")
	prefix := flag.String("prefix", "", "The prefix to search for objects to delete")
	maxItems := flag.Int64("items-per-page", 999, "The maximum number of items to return per page")
	flag.Parse()

	err := svc.ListObjectVersionsPages(&s3.ListObjectVersionsInput{
		Bucket:  bucket,
		Prefix:  prefix,
		MaxKeys: maxItems,
	}, DeletePage)

	if err != nil {
		panic(err)
	}
}
