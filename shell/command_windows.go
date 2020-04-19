//+build windows

package shell

// In Windows, all hierarchy will receive the signal (which is good because we cannot send it anyway)
// In fact, there's nothing for us to do here
func (c *Command) propagateSignal(pid int) {
	return
}
