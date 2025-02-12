# OKOne K8s Deployment Config

Major changes: 
- Add container command
- Add ports
- Add required env
- Add volume mount for config files
- Remove initContainer skywalking since it's not needed

## Sequencer & RPC
- Add volume mount for persistent data

```yaml
spec:
    # ...
    spec:
      volumes:
        - name: data
          persistentVolumeClaim:
            claimName: defi-xlayer-egseq-pvc

        # ...
        containers:
            # ...
            volumeMounts:
                - name: data
                mountPath: /home/erigon
```

- Set read/write permission for the volume
```yaml
    securityContext:
        fsGroup: 2000
```