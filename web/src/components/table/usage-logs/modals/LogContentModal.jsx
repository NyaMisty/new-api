/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useState } from 'react';
import { Modal, Button, Typography, Spin, Toast, Empty } from '@douyinfe/semi-ui';
import { API } from '../../../../helpers';

const { Title, Paragraph } = Typography;

const LogContentModal = ({ visible, onClose, logId, t }) => {
  const [loading, setLoading] = useState(false);
  const [contentData, setContentData] = useState({ request: null, response: null });

  React.useEffect(() => {
    if (visible && logId) {
      loadLogContent();
    }
  }, [visible, logId]);

  const normalizeContent = (data) => ({
    request: data?.request_body ?? data?.request ?? null,
    response: data?.response_body ?? data?.response ?? null,
  });

  const loadLogContent = async () => {
    setLoading(true);
    try {
      const res = await API.get(`/api/log/content/${logId}`);
      const { success, data, message } = res.data;
      if (success) {
        setContentData(normalizeContent(data));
      } else {
        Toast.error(message || t('加载失败'));
        onClose();
      }
    } catch (error) {
      Toast.error(t('加载失败: ') + error.message);
      onClose();
    } finally {
      setLoading(false);
    }
  };

  const parseContent = (payload) => {
    if (payload === null || payload === undefined) return '';
    if (typeof payload !== 'string') return payload;
    try {
      return JSON.parse(payload);
    } catch (e) {
      return payload;
    }
  };

  const hasRequest =
    contentData?.request !== null &&
    contentData?.request !== undefined &&
    !(typeof contentData.request === 'string' && contentData.request.trim() === '') &&
    !(typeof contentData.request === 'object' &&
      !Array.isArray(contentData.request) &&
      Object.keys(contentData.request).length === 0);

  const hasResponse =
    contentData?.response !== null &&
    contentData?.response !== undefined &&
    !(typeof contentData.response === 'string' && contentData.response.trim() === '') &&
    !(typeof contentData.response === 'object' &&
      !Array.isArray(contentData.response) &&
      Object.keys(contentData.response).length === 0);

  const hasContent = hasRequest || hasResponse;

  return (
    <Modal
      title={t('日志详情')}
      visible={visible}
      onCancel={onClose}
      footer={null}
      width={900}
      style={{ maxHeight: '80vh' }}
      bodyStyle={{ maxHeight: 'calc(80vh - 100px)', overflow: 'auto' }}
    >
      {loading ? (
        <div style={{ textAlign: 'center', padding: '50px 0' }}>
          <Spin size='large' />
        </div>
      ) : contentData ? (
        <div>
          {hasRequest && (
            <div style={{ marginBottom: 24 }}>
              <Title heading={5}>{t('请求内容')}</Title>
              <div
                style={{
                  background: '#f7f7f7',
                  padding: 16,
                  borderRadius: 4,
                  maxHeight: '300px',
                  overflow: 'auto',
                }}
              >
                <pre
                  style={{
                    margin: 0,
                    fontFamily:
                      'ui-monospace, SFMono-Regular, SF Mono, Menlo, Consolas, monospace',
                    fontSize: 13,
                    lineHeight: 1.45,
                    whiteSpace: 'pre-wrap',
                    wordBreak: 'break-word',
                  }}
                >
                  {typeof parseContent(contentData.request) === 'object'
                    ? JSON.stringify(parseContent(contentData.request), null, 2)
                    : contentData.request}
                </pre>
              </div>
            </div>
          )}

          {hasResponse && (
            <div>
              <Title heading={5}>{t('响应内容')}</Title>
              <div
                style={{
                  background: '#f7f7f7',
                  padding: 16,
                  borderRadius: 4,
                  maxHeight: '300px',
                  overflow: 'auto',
                }}
              >
                <pre
                  style={{
                    margin: 0,
                    fontFamily:
                      'ui-monospace, SFMono-Regular, SF Mono, Menlo, Consolas, monospace',
                    fontSize: 13,
                    lineHeight: 1.45,
                    whiteSpace: 'pre-wrap',
                    wordBreak: 'break-word',
                  }}
                >
                  {typeof parseContent(contentData.response) === 'object'
                    ? JSON.stringify(parseContent(contentData.response), null, 2)
                    : contentData.response}
                </pre>
              </div>
            </div>
          )}

          {!hasContent && (
            <div style={{ padding: '32px 0', textAlign: 'center' }}>
              <Empty
                image='https://lf3-static.bytednsdoc.com/obj/eden-cn/nasbopg/lwoqjlwzjjlkuljljkb/empty-image.png'
                description={<Paragraph>{t('暂无内容记录')}</Paragraph>}
              />
            </div>
          )}
        </div>
      ) : null}
    </Modal>
  );
};

export default LogContentModal;
