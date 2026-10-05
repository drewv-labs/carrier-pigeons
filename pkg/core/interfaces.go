package core

// TelemetryPublisher defines the behavior required to transmit CTD payloads.
// Any struct that has a Publish(*CTDPayload) error method automatically satisfies this.
type TelemetryPublisher interface {
	Publish(event *CTDPayload) error
}
