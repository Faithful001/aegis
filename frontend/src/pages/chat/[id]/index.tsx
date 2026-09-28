import React, { useEffect } from 'react';
import HomePage from '../../index';
import { useAegis } from '../../../context/AegisContext';

interface Props {
  id?: string;
}

export const DynamicChatThreadPage: React.FC<Props> = ({ id }) => {
  const { setActiveThreadId } = useAegis();

  useEffect(() => {
    if (id) {
      setActiveThreadId(id);
    }
  }, [id, setActiveThreadId]);

  return <HomePage />;
};

export default DynamicChatThreadPage;
