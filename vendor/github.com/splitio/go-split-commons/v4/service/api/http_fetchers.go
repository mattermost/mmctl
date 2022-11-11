package api

import (
	"bytes"
	"encoding/json"
	"net/url"

	"github.com/splitio/go-split-commons/v4/conf"
	"github.com/splitio/go-split-commons/v4/dtos"
	"github.com/splitio/go-split-commons/v4/service"
	"github.com/splitio/go-toolkit/v5/logging"
)

type httpFetcherBase struct {
	client Client
	logger logging.LoggerInterface
}

func (h *httpFetcherBase) fetchRaw(endpoint string, since int64, fetchOptions *service.FetchOptions) ([]byte, error) {
	queryParams, headers := service.BuildFetch(since, fetchOptions)
	if len(queryParams) > 0 {
		params := url.Values{}
		for key, value := range queryParams {
			params.Add(key, value)
		}
		endpoint = endpoint + "?" + params.Encode()
	}

	data, err := h.client.Get(endpoint, headers)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// HTTPSplitFetcher struct is responsible for fetching splits from the backend via HTTP protocol
type HTTPSplitFetcher struct {
	httpFetcherBase
}

// NewHTTPSplitFetcher instantiates and return an HTTPSplitFetcher
func NewHTTPSplitFetcher(apikey string, cfg conf.AdvancedConfig, logger logging.LoggerInterface, metadata dtos.Metadata) service.SplitFetcher {
	return &HTTPSplitFetcher{
		httpFetcherBase: httpFetcherBase{
			client: NewHTTPClient(apikey, cfg, cfg.SdkURL, logger, metadata),
			logger: logger,
		},
	}
}

// Fetch makes an http call to the split backend and returns the list of updated splits
func (f *HTTPSplitFetcher) Fetch(since int64, fetchOptions *service.FetchOptions) (*dtos.SplitChangesDTO, error) {
	data, err := f.fetchRaw("/splitChanges", since, fetchOptions)
	if err != nil {
		f.logger.Error("Error fetching split changes ", err)
		return nil, err
	}

	var splitChangesDto dtos.SplitChangesDTO
	err = json.Unmarshal(data, &splitChangesDto)
	if err != nil {
		f.logger.Error("Error parsing split changes JSON ", err)
		return nil, err
	}

	return &splitChangesDto, nil
}

// HTTPSegmentFetcher struct is responsible for fetching segment by name from the API via HTTP method
type HTTPSegmentFetcher struct {
	httpFetcherBase
}

// NewHTTPSegmentFetcher instantiates and returns a new HTTPSegmentFetcher.
func NewHTTPSegmentFetcher(apikey string, cfg conf.AdvancedConfig, logger logging.LoggerInterface, metadata dtos.Metadata) service.SegmentFetcher {
	return &HTTPSegmentFetcher{
		httpFetcherBase: httpFetcherBase{
			client: NewHTTPClient(apikey, cfg, cfg.SdkURL, logger, metadata),
			logger: logger,
		},
	}
}

// Fetch issues a GET request to the split backend and returns the contents of a particular segment
func (f *HTTPSegmentFetcher) Fetch(segmentName string, since int64, fetchOptions *service.FetchOptions) (*dtos.SegmentChangesDTO, error) {
	var bufferQuery bytes.Buffer
	bufferQuery.WriteString("/segmentChanges/")
	bufferQuery.WriteString(segmentName)

	data, err := f.fetchRaw(bufferQuery.String(), since, fetchOptions)
	if err != nil {
		f.logger.Error(err.Error())
		return nil, err
	}
	var segmentChangesDto dtos.SegmentChangesDTO
	err = json.Unmarshal(data, &segmentChangesDto)
	if err != nil {
		f.logger.Error("Error parsing segment changes JSON for segment ", segmentName, err)
		return nil, err
	}

	return &segmentChangesDto, nil
}
