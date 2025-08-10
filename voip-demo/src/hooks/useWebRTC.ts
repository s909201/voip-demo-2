import { useState, useRef, useCallback } from 'react';

type CallState = 'idle' | 'calling' | 'in-call' | 'ended';
type UploadStatus = 'idle' | 'uploading' | 'success' | 'error';

const ICE_SERVERS = {
  iceServers: [
    { urls: 'stun:stun.l.google.com:19302' },
  ],
};

export const useWebRTC = () => {
  const [callState, setCallState] = useState<CallState>('idle');
  const [remoteStream, setRemoteStream] = useState<MediaStream | null>(null);
  const [uploadStatus, setUploadStatus] = useState<UploadStatus>('idle');
  const peerConnection = useRef<RTCPeerConnection | null>(null);
  const localStream = useRef<MediaStream | null>(null);
  const mediaRecorder = useRef<MediaRecorder | null>(null);
  const audioChunks = useRef<Blob[]>([]);

  const uploadAudio = useCallback(async (audioBlob: Blob, callId?: string) => {
    try {
      console.log('開始上傳音訊檔案...');
      
      // 創建 FormData 物件
      const formData = new FormData();
      formData.append('audioFile', audioBlob, 'recording.wav');
      
      // 可選擇性地附加其他元數據
      if (callId) {
        formData.append('callId', callId);
      }
      formData.append('timestamp', new Date().toISOString());
      
      // 發送 POST 請求到後端
      const response = await fetch('/api/upload', {
        method: 'POST',
        body: formData,
      });
      
      if (!response.ok) {
        throw new Error(`上傳失敗: ${response.status} ${response.statusText}`);
      }
      
      const result = await response.json();
      console.log('音訊檔案上傳成功:', result);
      return result;
      
    } catch (error) {
      console.error('音訊檔案上傳失敗:', error);
      throw error;
    }
  }, []);

  const openUserMedia = useCallback(async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true, video: false });
      localStream.current = stream;
    } catch (error) {
      console.error("Error accessing media devices.", error);
    }
  }, []);

  const initializePeerConnection = useCallback((send?: (data: string) => void) => {
    const pc = new RTCPeerConnection(ICE_SERVERS);

    pc.ontrack = (event) => {
      const stream = event.streams[0];
      setRemoteStream(stream);
      
      // 初始化 MediaRecorder 來錄製遠端音訊流
      try {
        const recorder = new MediaRecorder(stream, {
          mimeType: 'audio/webm;codecs=opus'
        });
        
        // 綁定 ondataavailable 事件監聽器
        recorder.ondataavailable = (event) => {
          if (event.data.size > 0) {
            audioChunks.current.push(event.data);
            console.log('音訊數據塊已收集，大小:', event.data.size);
          }
        };

        // 綁定 onstop 事件監聽器
        recorder.onstop = async () => {
          console.log('錄音已停止，處理音訊數據...');
          
          if (audioChunks.current.length > 0) {
            // 將收集到的數據塊合併成一個 Blob 物件
            const audioBlob = new Blob(audioChunks.current, { type: 'audio/wav' });
            console.log('音訊 Blob 已生成，大小:', audioBlob.size, '位元組');
            
            // 清空音訊數據塊陣列，為下一次錄音做準備
            audioChunks.current = [];
            
            // 開始上傳流程
            setUploadStatus('uploading');
            
            try {
              await uploadAudio(audioBlob);
              setUploadStatus('success');
              console.log('音訊上傳成功');
            } catch (error) {
              setUploadStatus('error');
              console.error('音訊上傳失敗:', error);
            }
          } else {
            console.warn('沒有音訊數據可處理');
          }
        };
        
        mediaRecorder.current = recorder;
        console.log('MediaRecorder 已初始化');
      } catch (error) {
        console.error('初始化 MediaRecorder 失敗:', error);
      }
    };

    pc.onicecandidate = (event) => {
      if (event.candidate && send) {
        const candidateMessage = {
          type: 'candidate',
          payload: event.candidate,
        };
        send(JSON.stringify(candidateMessage));
      }
    };

    peerConnection.current = pc;
  }, []);

  const startCall = useCallback(async (targetUserId: string, send: (data: string) => void) => {
    initializePeerConnection(send);
    await openUserMedia();
    if (peerConnection.current && localStream.current) {
      localStream.current.getTracks().forEach(track => {
        peerConnection.current?.addTrack(track, localStream.current!);
      });

      const offer = await peerConnection.current.createOffer();
      await peerConnection.current.setLocalDescription(offer);

      const offerMessage = {
        type: 'offer',
        targetUserId: targetUserId,
        payload: offer,
      };
      send(JSON.stringify(offerMessage));
      setCallState('calling');
    }
  }, [initializePeerConnection, openUserMedia]);

  const handleOffer = useCallback(async (offer: RTCSessionDescriptionInit, callerId: string, send: (data: string) => void) => {
    initializePeerConnection(send);
    await openUserMedia();
    if (peerConnection.current && localStream.current) {
      await peerConnection.current.setRemoteDescription(new RTCSessionDescription(offer));
      
      localStream.current.getTracks().forEach(track => {
        peerConnection.current?.addTrack(track, localStream.current!);
      });

      const answer = await peerConnection.current.createAnswer();
      await peerConnection.current.setLocalDescription(answer);

      const answerMessage = {
        type: 'answer',
        targetUserId: callerId,
        payload: answer,
      };
      send(JSON.stringify(answerMessage));
      setCallState('in-call');
    }
  }, [initializePeerConnection, openUserMedia]);

  const handleAnswer = useCallback(async (answer: RTCSessionDescriptionInit) => {
    if (peerConnection.current) {
      await peerConnection.current.setRemoteDescription(new RTCSessionDescription(answer));
      setCallState('in-call');
    }
  }, []);

  const handleCandidate = useCallback(async (candidate: RTCIceCandidateInit) => {
    if (peerConnection.current) {
      await peerConnection.current.addIceCandidate(new RTCIceCandidate(candidate));
    }
  }, []);

  const startRecording = useCallback(() => {
    if (mediaRecorder.current && mediaRecorder.current.state === 'inactive') {
      // 清空之前的音訊數據塊
      audioChunks.current = [];
      mediaRecorder.current.start();
      console.log('開始錄音');
    } else {
      console.warn('MediaRecorder 未準備好或已在錄音中');
    }
  }, []);

  const stopRecording = useCallback(() => {
    if (mediaRecorder.current && mediaRecorder.current.state === 'recording') {
      mediaRecorder.current.stop();
      console.log('停止錄音');
    } else {
      console.warn('MediaRecorder 未在錄音中');
    }
  }, []);

  const resetUploadStatus = useCallback(() => {
    setUploadStatus('idle');
  }, []);

  const hangUp = useCallback(() => {
    // 如果正在錄音，先停止錄音
    if (mediaRecorder.current && mediaRecorder.current.state === 'recording') {
      stopRecording();
    }
    
    if (peerConnection.current) {
      peerConnection.current.close();
      peerConnection.current = null;
    }
    if (localStream.current) {
      localStream.current.getTracks().forEach(track => track.stop());
      localStream.current = null;
    }
    setRemoteStream(null);
    setCallState('ended');
    // 重置上傳狀態
    setUploadStatus('idle');
    // 可選：發送一個 'hangup' 信令通知對方
  }, [stopRecording]);

  return {
    callState,
    uploadStatus,
    localStream: localStream.current,
    remoteStream,
    mediaRecorder: mediaRecorder.current,
    audioChunks: audioChunks.current,
    openUserMedia,
    initializePeerConnection,
    startCall,
    handleOffer,
    handleAnswer,
    handleCandidate,
    hangUp,
    startRecording,
    stopRecording,
    uploadAudio,
    resetUploadStatus,
  };
};
