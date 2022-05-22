package iso9660

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMangleDirectoryName(t *testing.T) {
	for _, testcase := range []struct {
		input  string
		output string
	}{
		{
			input:  "ThisStringIsFarTooLongToBeWritten",
			output: "thisstringisfartoolongtobewritt",
		},
		{
			input:  "ThisStringHasUnicodeCharacterŁ",
			output: "thisstringhasunicodecharacter__",
		},
		{
			input:  "ThisStringHasItByteBeforeThEndŁ",
			output: "thisstringhasitbytebeforethend_",
		},
	} {
		t.Run(testcase.input, func(t *testing.T) {
			o := mangleDirectoryName(testcase.input)
			assert.Equal(t, testcase.output, o)
		})
	}
}

func TestMangleFileName(t *testing.T) {
	for _, testcase := range []struct {
		input  string
		output string
	}{
		{
			input:  "ThisStringIsFarTooLongToBeWritten",
			output: "thisstringisfartoolongtobewr;1",
		},
		{
			input:  "ThisStringHasUnicodeCharacŁ",
			output: "thisstringhasunicodecharac__;1",
		},
		{
			input:  "ThisStringHasAFileExtensionAndItIsVery.Long",
			output: "thisstringhasafileexten.long;1",
		},
		{
			input:  "ThisStringHasAFileExtensionThats.FarTooLong",
			output: "thisstringhasafilee.fartoolo;1",
		},
	} {
		t.Run(testcase.input, func(t *testing.T) {
			o := mangleFileName(testcase.input)
			assert.Equal(t, testcase.output, o)
		})
	}
}

func TestWriterStaging(t *testing.T) {
	w, err := NewWriter()
	assert.NoError(t, err)
	defer w.Cleanup() // nolint: errcheck

	testFileContents := "hrh2309hr320h"
	testFilePath := "FarTooLongFilePathThatWillBeTrimmed/dirø1/somefile.dat"
	testFileMangledPath := "fartoolongfilepaththatwillbetri/dir__1/somefile.dat;1"

	r := strings.NewReader(testFileContents)
	err = w.AddFile(r, testFilePath)
	assert.NoError(t, err)

	readData, err := ioutil.ReadFile(path.Join(w.stagingDir, testFileMangledPath))
	assert.NoError(t, err)

	assert.Equal(t, testFileContents, string(readData))
}

func TestWriter(t *testing.T) {
	w, err := NewWriter()
	assert.NoError(t, err)
	defer func() {
		if cleanupErr := w.Cleanup(); cleanupErr != nil {
			t.Fatalf("failed to cleanup writer: %v", cleanupErr)
		}
	}()

	err = w.AddFile(strings.NewReader("hrh2309hr320h"), "someDirectory-Path/dir1/somefile.dat")
	assert.NoError(t, err)

	largeFileData, err := ioutil.ReadFile("fixtures/test.iso_source/dir2/large.txt")
	assert.NoError(t, err)

	err = w.AddLocalFile("fixtures/test.iso_source/dir2/large.txt", "anotherDir/large.txt")
	assert.NoError(t, err)

	f, err := ioutil.TempFile(os.TempDir(), "iso9660_golang_test")
	assert.NoError(t, err)
	defer os.Remove(f.Name())

	imageFileName := f.Name()

	err = w.WriteTo(f, "testvolume")
	assert.NoError(t, err)

	f.Close() // nolint: errcheck

	f, err = os.Open(imageFileName)
	assert.NoError(t, err)

	img, err := OpenImage(f)
	assert.NoError(t, err)

	root, err := img.RootDir()
	assert.NoError(t, err)

	children, err := root.GetChildren()
	assert.NoError(t, err)
	assert.Len(t, children, 2)
	assert.Equal(t, "anotherdir", children[0].Name())

	children, err = children[0].GetChildren()
	assert.NoError(t, err)
	assert.Len(t, children, 1)
	assert.Equal(t, "large.txt", children[0].Name())

	readData, err := ioutil.ReadAll(children[0].Reader())
	assert.NoError(t, err)

	assert.Equal(t, largeFileData, readData)
}
