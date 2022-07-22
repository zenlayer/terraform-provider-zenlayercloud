package connectivity

import (
        "github.com/zenlayer/zenlayer-go-sdk/services/bmc"
)

type ZenlayerClient struct {
        AccessKeyId string
        AccessKeyPassword string
}

func (client *ZenlayerClient) NewBmcClient() (*bmc.Client, error){
        return bmc.NewClientWithAccessKey(client.AccessKeyId, client.AccessKeyPassword)
}


