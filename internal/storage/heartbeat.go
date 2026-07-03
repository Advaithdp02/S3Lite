package storage

import (
	"os"
	"time"
)

func (s* Storage) StartHeartBeat(interval time.Duration){
	go func(){
		ticker:=time.NewTicker(interval)
		defer ticker.Stop()
		
		for{
			for i:=range s.Nodes{
				node:= &s.Nodes[i]
				_,err:=os.Stat(node.Path)
				node.mu.Lock()
				node.Alive=err==nil
				node.LastHeartbeat=time.Now()

				node.mu.Unlock()



			}
			<-ticker.C
		}
	}()
}
func (n *Node) IsAlive() bool {

	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.Alive
}
